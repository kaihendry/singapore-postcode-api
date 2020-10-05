<script src="https://embed.github.com/view/geojson/kaihendry/singapore-postcode/master/eg.geojson"></script>

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
