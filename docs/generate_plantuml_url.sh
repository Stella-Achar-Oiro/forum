#!/bin/bash

# Read the PlantUML content
content=$(cat docs/database_erd.puml)

# URL encode the content
encoded=$(echo -n "$content" | xxd -p | tr -d '\n' | sed 's/\(..\)/%\1/g')

# Generate the PlantUML server URL
echo "http://www.plantuml.com/plantuml/png/$encoded" 