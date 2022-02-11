#!/bin/sh

# Share an album as "Anyone with link can view" in Synology Photos
SHARELINK='https://b92.dsmdemo.synologydemo.com:5001/mo/sharing/k5SnJvlVW'
DIR="$(dirname $0)"

cookie=$(curl -k "${SHARELINK}" -H 'User-Agent: Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:93.0) Gecko/20100101 Firefox/93.0' -H 'Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8' -H 'Accept-Language: en-US,en;q=0.5' --compressed -H 'DNT: 1' -H 'Connection: keep-alive' -H 'Upgrade-Insecure-Requests: 1' -H 'Sec-Fetch-Dest: document' -H 'Sec-Fetch-Mode: navigate' -H 'Sec-Fetch-Site: none' -H 'Sec-Fetch-User: ?1' -i | grep Cookie)
cookie="${cookie#*:}"
cookie="${cookie%;*}"
base_url="${SHARELINK%/*}"
album_code="${SHARELINK##*/}"

album=$(curl -k "${base_url}/webapi/entry.cgi?" -X POST \
-H "x-syno-sharing: ${album_code}" \
-H "Cookie: ${cookie}" \
-d 'offset=0&limit=1000&api="SYNO.Foto.Browse.Item"&method="list"&version=1 -b cookies.txt')

length=$(($(echo "${album}" | jq -r '.data[] | length ')-1))
random=$(shuf -i 1-${length} -n 1)
id=$(echo "${album}" | jq -r ".data[] | .[${random}] | .id ")

if [ ! -f "${DIR}/../cache/${id}.png" ]; then
    curl -k "${base_url}/webapi/entry.cgi/20210807_144336.jpg" -G \
    -H "Cookie: ${cookie}" \
    -d "id=${id}" -d "cache_key=\"35336_1628372812\"&type=\"unit\"&size=\"xl\"&passphrase=\"${album_code}\"&api=\"SYNO.Foto.Thumbnail\"&method=\"get\"&version=1&_sharing_id=\"${album_code}\"" \
    -o "${DIR}/../cache/unprocessed.jpg"
    "${DIR}/../bin/convert" "${DIR}/../cache/unprocessed.jpg" -filter LanczosSharp -brightness-contrast 3x15 -resize 1448x -gravity center -crop 1448x1072+0+0 +repage -rotate 90 -colorspace Gray -dither FloydSteinberg -remap "${DIR}/../bin/kindle_colors.gif" -quality 75 -define png:color-type=0 -define png:bit-depth=8 "${DIR}/../cache/${id}.png"
fi

cp "${DIR}/../cache/${id}.png" "${DIR}/../cache/current_photo.png"


