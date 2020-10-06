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

This advanced example puts all original values in properties which should be
viewable by viewing over the map point:

	#!/bin/bash
	cat << EOF |
	[ { "foo": 42, "something": "else", "postcode": "069535" }, { "foo": 12, "more": { "age": 12}, "postcode": "439870" } ]
	EOF
	jq -c -r '.[]' | while read -r json
	do
		jq -r '.postcode' <<< $json | while read -r postcode
		do
			curl -s "https://postcode.dabase.com/?postcode=$postcode" |
				jq --argjson props "$json" '.properties |= $props'
		done
	done | jq -s '{ "type": "FeatureCollection", "features": . }'

h/t geirha in #jq
