#!/bin/bash

# Minify the JavaScript file
terser index.js -c -m -o index.min.js

# Read the minified JavaScript file content into a variable
JS_CONTENT=$(<index.min.js)

# Create a temporary file for our new content
TMP_FILE=$(mktemp)

# Create a copy of index.html with the new JavaScript content
while IFS= read -r line; do
  if [[ "$line" =~ \/\/MINIFIED ]]; then
    printf '%s' "$JS_CONTENT" >> "$TMP_FILE"
  else
    printf '%s\n' "$line" >> "$TMP_FILE"
  fi
done < customerInsights.html

# Now, use grep to create the final public.html file.
grep -v '<script src="./index.js"></script>' "$TMP_FILE" > customerInsights.min.html

# Remove the temporary file
rm "$TMP_FILE"
rm index.min.js

echo "Done..."
