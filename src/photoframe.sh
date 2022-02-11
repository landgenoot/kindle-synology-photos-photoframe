DEBUG=${DEBUG:-false}
[ $DEBUG = true ] && set -x

DIR="$(dirname $0)"
PHOTO_PNG="${DIR}/../cache/current_photo.png"
FETCH_DASHBOARD_CMD="$DIR/refresh.sh"
LOW_BATTERY_CMD="$DIR/local/low-battery.sh"

REFRESH_SCHEDULE="0 6 * * *"
TIMEZONE="Europe/Amsterdam"
WIFI_TEST_IP="1.1.1.1"
FULL_DISPLAY_REFRESH_RATE=${FULL_DISPLAY_REFRESH_RATE:-0}
SLEEP_SCREEN_INTERVAL=${SLEEP_SCREEN_INTERVAL:-3600}
RTC=/sys/devices/platform/mxc_rtc.0/wakeup_enable

LOW_BATTERY_REPORTING=${LOW_BATTERY_REPORTING:-false}
LOW_BATTERY_THRESHOLD_PERCENT=${LOW_BATTERY_THRESHOLD_PERCENT:-10}

num_refresh=0

init() {
  echo "Starting dashboard with $REFRESH_SCHEDULE refresh..."
  /etc/upstart/framework stop
  initctl stop webreader > /dev/null 2>&1
  echo powersave > /sys/devices/system/cpu/cpu0/cpufreq/scaling_governor
  lipc-set-prop com.lab126.powerd preventScreenSaver 1
  echo -n 0 > /sys/class/backlight/max77696-bl/brightness
}

refresh_dashboard() {
  echo "Refreshing dashboard"
  if "$DIR/wait-for-wifi.sh" "$WIFI_TEST_IP"; then
	  "$FETCH_DASHBOARD_CMD"

	  if [ $num_refresh -eq $FULL_DISPLAY_REFRESH_RATE ]; then
      num_refresh=0

      # trigger a full refresh once in every 4 refreshes, to keep the screen clean
      echo "Full screen refresh"
      /usr/sbin/eips -f -g "$PHOTO_PNG"
	  else
      echo "Partial screen refresh"
      /usr/sbin/eips -g "$PHOTO_PNG"
	  fi

	  num_refresh=$((num_refresh+1))
  else
    echo Connection error, skipping refresh
  fi
}

log_battery_stats() {
  battery_level=$(gasgauge-info -c)
  echo "$(date) Battery level: $battery_level."

  if [ $LOW_BATTERY_REPORTING = true ]; then
    battery_level_numeric=${battery_level%?}
    if [ $battery_level_numeric -le $LOW_BATTERY_THRESHOLD_PERCENT ]; then
      "$LOW_BATTERY_CMD" $battery_level_numeric
    fi
  fi
}

rtc_sleep() {
  duration=$1

  if [ $DEBUG = true ]; then
    sleep $duration
  else
    echo "" > /sys/class/rtc/rtc1/wakealarm
    # Following line contains the sleep time in seconds
    echo "+${duration}" > /sys/class/rtc/rtc1/wakealarm
    echo "mem" > /sys/power/state
  fi
}

main_loop() {
  while true ; do
    log_battery_stats

    next_wakeup_secs=$("${DIR}/../bin/next-wakeup" --schedule="$REFRESH_SCHEDULE" --timezone="$TIMEZONE")

    action="suspend"
    refresh_dashboard

    # take a bit of time before going to sleep, so this process can be aborted
    sleep 10

    echo "Going to $action, next wakeup in ${next_wakeup_secs}s"

    #rtc_sleep $next_wakeup_secs
  done
}

init
main_loop
