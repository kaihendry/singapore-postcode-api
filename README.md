# Singapore postcode geocoder

	#!/bin/bash
	cat << EOF |
	[
	{ "name": "Office", "postcode": "069535" },
	{ "name": "Bar", "postcode": "238875" }
	]
	EOF
	jq -r '.[]|.name,.postcode' |
	while read -r name
	read -r postcode; do
	curl -s "https://postcode.dabase.com/?postcode=$postcode" |
		jq --arg name $name '.properties.Name = $name'
	done | jq -s '{ "type": "FeatureCollection", "features": . }'

Example [GeoJSON output](https://docs.github.com/en/free-pro-team@latest/github/managing-files-in-a-repository/mapping-geojson-files-on-github) https://github.com/kaihendry/singapore-postcode/blob/master/eg.geojson
