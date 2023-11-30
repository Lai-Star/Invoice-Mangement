#!/usr/bin/env bash

# Try to hit the ngrok API inside docker compose, if this succeeds then that means webhooks are enabled for plaid for
# local development.
WEBHOOKS_DOMAIN=$(curl http://ngrok:4040/api/tunnels -s -m 0.1 | perl -pe '/\"public_url\":\"https:\/\/(\S*?)\",/g; print $1;' | cut -d "{" -f1);

if [[ ! -z "${WEBHOOKS_DOMAIN}" ]]; then
  echo "[wrapper] ngrok detected, webhooks should target: ${WEBHOOKS_DOMAIN}";

  # If the domain name has been derived then enable webhooks for plaid.
  echo "[wrapper] Plaid webhooks have been enabled...";
  export MONETR_PLAID_WEBHOOKS_DOMAIN=${WEBHOOKS_DOMAIN};
  export MONETR_PLAID_WEBHOOKS_ENABLED="true";
else
  echo "[wrapper] ngrok not detected, webhooks will not be available..."
fi

# If the stripe API key, webhook secret and price ID are provided then enable billing for local development.
# Stripe does require webhooks, as we rely on them in order to know when a subscription becomes active.
if [[ ! -z "${MONETR_STRIPE_API_KEY}" ]] && \
  [[ ! -z "${MONETR_STRIPE_WEBHOOK_SECRET}" ]] && \
  [[ ! -z "${MONETR_STRIPE_DEFAULT_PRICE_ID}" ]] && \
  [[ ! -z "${WEBHOOKS_DOMAIN}" ]]; then
  echo "[wrapper] Stripe credentials are available, stripe and billing will be enabled...";
  export MONETR_STRIPE_ENABLED="true";
  export MONETR_STRIPE_BILLING_ENABLED="true";
  export MONETR_STRIPE_WEBHOOKS_ENABLED="true";
  export MONETR_STRIPE_WEBHOOKS_DOMAIN=${WEBHOOKS_DOMAIN};
fi

# Execute the command with the new environment variables.
/go/bin/dlv exec --continue --accept-multiclient --listen=:2345 --headless=true --api-version=2 /usr/bin/monetr -- serve --migrate=true;
