#!/bin/bash

BASE_URL="http://localhost:8080"

function upload_file() {
    local file_path="$1"
    if [ ! -f "$file_path" ]; then
        echo "File not found: $file_path"
        exit 1
    fi

    echo "Uploading $file_path..."

    response=$(curl -s -D - -F "file=@${file_path}" "$BASE_URL/api/upload") 

    location_header_line=$(echo "${response%%$'\r\n\r\n'*}" | grep -i "^Location:")
    body="${response#*$'\r\n\r\n'}"

    if [ -z "$location_header_line" ]; then
        echo -e "Failed uploading file"
        echo "$response"
        exit 1
    fi

    location=$(echo "$location_header_line" | cut -d' ' -f2- | cut -c6-)

    echo -e "$BASE_URL$location"
    exit 0
}

if [ "$1" == "upload" ]; then
    upload_file "$2"
elif [ "$1" == "download" ]; then
    download_file "$2" "$3"
else
    echo "Invalid command. Use 'upload <file>' or 'download <userId> <fileId>'"
    exit 1
fi
