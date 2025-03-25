# kindle-synology-photos-photoframe

Display a random photo on a jailbroken Kindle every morning, either from a shared Synology Photos album **or** from a local folder with image files.  
All image processing happens locally on the device using 100% Golang and ImageMagick.

Read the corresponding blog post [here](https://daanmiddendorp.com/tech/2022/02/14/new-destination-for-my-broken-kindle)

![Photo frame](https://daanmiddendorp.com/assets/responsive-images/895/20220214_151832.jpg)

Tested on a Kindle Voyage. Should also work on other jailbroken Kindles.

---

## ✅ Features

- 🖼️ Load photos from a **Synology Photos public album** _or_ a **local folder**
- 🧠 Converts images to grayscale Kindle-compatible format with dithering and contrast adjustments
- 🔋 Displays battery level and shows low battery icon if needed
- 💤 Suspends Kindle and wakes up automatically at your defined time

---

## 📦 Requirements

- Jailbroken Kindle that supports SSH  
  ([<= 5.13.3](https://www.mobileread.com/forums/showthread.php?t=338268) or [<= 5.14.2](https://www.mobileread.com/forums/showthread.php?t=346037))
- Golang toolchain (for building from source)
- ImageMagick development libraries (e.g., `libmagickwand-dev` on Debian)

---

## 🚀 Installation

1. Jailbreak your Kindle and enable SSH access
2. Download the [latest release](https://github.com/landgenoot/kindle-synology-photos-photoframe/releases/latest)
3. Copy the `photoframe` binary and `linkss` folder to the Kindle internal storage at `/mnt/us`
4. Launch the app using one of the following modes:

### 🔗 Synology NAS Album Mode

```bash
./photoframe http://192.168.50.57:5000/mo/sharing/RMVJ3g6t8 > /dev/null &
```
Replace the URL with your own Synology shared album link

Ensure album sharing settings are set to:

- "Public – Anyone with the link can view"

## 📁 Local Folder Mode

```bash
./photoframe /mnt/us/photos > /dev/null &
```

Replace `/mnt/us/photos` with the path to your local photo folder (JPEG/PNG files supported)

## 🛑 Stopping

To stop the program:

```bash
killall photoframe
```
You’ll have 30 seconds to stop it after waking up the Kindle manually by pressing the power button.

## 🪵 Logs
Log file is stored at:

```bash
/mnt/us/photoframe.log
```