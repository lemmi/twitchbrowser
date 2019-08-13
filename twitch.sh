#!/bin/bash --debugger

listonline () {
	twitchbrowser -names "${@,,}" 
}

if [[ "$#" -eq "1" ]]; then
	explicit=true
fi

if [[ "$#" -eq "0" ]]; then
	set -- -fav
fi

stream=""
online=$(listonline "$@" | sort)
nonline=$(wc -l <<< "$online")

if [[ "$nonline" -eq "1" ]]; then
	echo "$online"
	stream="$online"
	if [[ ! "$explicit" == "true" ]]; then
		read || exit
	fi
elif [[ "$nonline" -gt "1" ]]; then
	PS3="Select Stream: "
	select choice in $online; do
		[[ "x$choice" != "x" ]] && stream="$choice"
		break
	done
fi

[[ -n "$stream" ]] && mpv --fs  https://twitch.tv/"$stream"
