
GOOS=darwin
GOARCHS=(
	amd64
	arm64
)

name="mac-airtag-to-mqtt"
for GOARCH in "${GOARCHS[@]}"; do
	echo "Building for $GOOS/$GOARCH"
	GOOS=$GOOS GOARCH=$GOARCH go build -ldflags="-s -w" -o "$name-$GOOS-$GOARCH" .
done
#lipo -create -output mac-airtag-to-mqtt mac-airtag-to-mqtt-darwin-*
