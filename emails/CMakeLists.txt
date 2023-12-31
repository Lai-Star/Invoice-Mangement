set(EMAIL_TEMPLATE_INPUT_DIRECTORY "${CMAKE_CURRENT_SOURCE_DIR}/src")
set(EMAIL_TEMPLATE_OUTPUT_DIRECTORY "${CMAKE_SOURCE_DIR}/server/communication/email_templates")

# WARNING: This does not exclude any spec or story files.
file(GLOB_RECURSE EMAIL_TEMPLATE_INPUTS "${EMAIL_TEMPLATE_INPUT_DIRECTORY}/*.tsx")

set(HTML_EMAIL_TEMPLATES "${EMAIL_TEMPLATE_INPUTS}")
list(TRANSFORM HTML_EMAIL_TEMPLATES REPLACE "${EMAIL_TEMPLATE_INPUT_DIRECTORY}" "${EMAIL_TEMPLATE_OUTPUT_DIRECTORY}")
list(TRANSFORM HTML_EMAIL_TEMPLATES REPLACE ".tsx" ".html")

set(TEXT_EMAIL_TEMPLATES "${EMAIL_TEMPLATE_INPUTS}")
list(TRANSFORM TEXT_EMAIL_TEMPLATES REPLACE "${EMAIL_TEMPLATE_INPUT_DIRECTORY}" "${EMAIL_TEMPLATE_OUTPUT_DIRECTORY}")
list(TRANSFORM TEXT_EMAIL_TEMPLATES REPLACE ".tsx" ".txt")

add_custom_command(
  OUTPUT ${HTML_EMAIL_TEMPLATES} ${TEXT_EMAIL_TEMPLATES}
  BYPRODUCTS ${HTML_EMAIL_TEMPLATES} ${TEXT_EMAIL_TEMPLATES}
  COMMAND ${REACT_EMAIL_EXECUTABLE} export --dir ${EMAIL_TEMPLATE_INPUT_DIRECTORY} --outDir ${EMAIL_TEMPLATE_OUTPUT_DIRECTORY}
  COMMAND ${REACT_EMAIL_EXECUTABLE} export --plainText --dir ${EMAIL_TEMPLATE_INPUT_DIRECTORY} --outDir ${EMAIL_TEMPLATE_OUTPUT_DIRECTORY}
  COMMENT "Building email templates from React"
  DEPENDS
    tools.react-email
    ${EMAIL_TEMPLATE_INPUTS}
)

add_custom_target(
  build.email
  COMMAND exit
  DEPENDS
    ${HTML_EMAIL_TEMPLATES}
    ${TEXT_EMAIL_TEMPLATES}
)

set(EMAIL_DEV_DIRECTORY ${CMAKE_BINARY_DIR}/development.email)
file(MAKE_DIRECTORY ${EMAIL_DEV_DIRECTORY})

add_custom_command(
  OUTPUT ${EMAIL_DEV_DIRECTORY}/package.json
  BYPRODUCTS ${EMAIL_DEV_DIRECTORY}/package.json
  COMMENT "Setting up required filed for react-email development"
  COMMAND ${CMAKE_COMMAND} -E create_symlink ${CMAKE_CURRENT_SOURCE_DIR}/package.json ${EMAIL_DEV_DIRECTORY}/package.json
)

add_custom_target(
  development.email
  COMMENT "Running the local development tools for email templates"
  COMMAND ${REACT_EMAIL_EXECUTABLE} dev --dir ${EMAIL_TEMPLATE_INPUT_DIRECTORY}
  WORKING_DIRECTORY ${EMAIL_DEV_DIRECTORY}
  DEPENDS
    tools.react-email
    ${EMAIL_DEV_DIRECTORY}/package.json
)
