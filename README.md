# mac_airtag_to_mqtt

This project is porting the original [airtag_to_mqtt](https://github.com/ndbroadbent/mac_airtag_to_mqtt) to Go, so that it's a bit easier to install.

Fetches AirTag data from `~/Library/Caches/com.apple.findmy.fmipcore/Items.data`, creates entities in Home Assistant with location data via [MQTT Discovery](https://www.home-assistant.io/integrations/mqtt/#mqtt-discovery).
