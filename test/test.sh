# Successfuly Generate
echo "Sending text 'some-text' to the service"
A=$(curl \
    -s \
    -X POST \
    -H "Content-Type: application/json" \
    -d '{"link": "some-text"}' http://localhost:3000/gen \
    2>&1)

# Shortened
SA=${A: -6}
echo "Result: " $A

# Get the shortened
echo "Requesting $A"
GA=$(curl \
    -s \
    "http://localhost:3000/lnk/${SA}" \
    2>&1)

echo "Result:" $GA

# Wrong request
echo "Asking for an invalid link..."
WR=$(curl \
    -s \
    "http://localhost:3000/lnk/blahbl" \
    2>&1)

echo "Result:" $WR
