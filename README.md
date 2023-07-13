# kindle-synology-photos-photoframe
Pick a random photo from a shared album in Synology Photos and show it on a jailbroken Kindle every morning.
All processing happens on the device using Golang and Imagemagick.

Read the corresponding blog post [here](https://daanmiddendorp.com/tech/2022/02/14/new-destination-for-my-broken-kindle)

![Photo frame](https://daanmiddendorp.com/assets/responsive-images/895/20220214_151832.jpg)

Tested on a Kindle Voyage. But should also work on other jailbroken Kindles.


## Instructions
1. Make sure your Kindle is jailbroken and is reachable over SSH ([< 5.13.3](https://www.mobileread.com/forums/showthread.php?t=338268) or [< 5.14.2](https://www.mobileread.com/forums/showthread.php?t=346037)
2. Download the [latest release](https://github.com/landgenoot/kindle-synology-photos-photoframe/releases/latest)
3. Copy `photoframe` binary and `linkss` folder to the internal storage of the kindle (`/mnt/us`).
4. To start `/mnt/us/photoframe http://192.168.50.57:5000/mo/sharing/RMVJ3g6t8 > /dev/null &` (replace url with sharing link to your own album)
5. To stop `killall photoframe`.

Log files are stored under `/mnt/us/photoframe.log` 
