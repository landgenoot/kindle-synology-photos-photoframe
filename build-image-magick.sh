#!/bin/sh

# This script build a tiny version of ImageMagick. Based on the NiLuje's hacks: http://www.mobileread.mobi/forums/showthread.php?t=225030

cd /home
curl https://imagemagick.org/archive/releases/ImageMagick-6.9.12-58.tar.xz -O 
tar xvJf ImageMagick-6.9.12-58.tar.xz 
cd ImageMagick-6.9.12-58

export ARCH_FLAGS="-march=armv7-a -mtune=cortex-a9 -mfpu=neon -mfloat-abi=softfp -mthumb"
export CROSS_TC="arm-linux-gnueabi"
export AUTO_JOBS=$(($(getconf _NPROCESSORS_ONLN 2> /dev/null || echo 0) + 1))
export JOBSFLAGS="-j${AUTO_JOBS}"
export BASE_CFLAGS="-O3 -ffast-math ${ARCH_FLAGS} ${LEGACY_GLIBCXX_ABI} -pipe -fomit-frame-pointer -frename-registers -fweb -flto=${AUTO_JOBS} -fuse-linker-plugin"
export NOLTO_CFLAGS="-O3 -ffast-math ${ARCH_FLAGS} ${LEGACY_GLIBCXX_ABI} -pipe -fomit-frame-pointer -frename-registers -fweb"
export RICE_CFLAGS="-O3 -ffast-math -ftree-vectorize -funroll-loops ${ARCH_FLAGS} ${LEGACY_GLIBCXX_ABI} -pipe -fomit-frame-pointer -frename-registers -fweb -flto=${AUTO_JOBS} -fuse-linker-plugin"
export CFLAGS="${RICE_CFLAGS}"
export DEVICE_USERSTORE="/mnt/us"
export LDFLAGS="${BASE_LDFLAGS} -Wl,-rpath=${DEVICE_USERSTORE}/linkss/lib"
export ax_cv_check_cl_libcl=no
export PKG_CONFIG="${BASE_PKG_CONFIG} --static"
export TC_BUILD_DIR="/home/ImageMagick-6.9.12-58/build"

env LIBS="-lrt" ./configure --prefix=${TC_BUILD_DIR}  --host=${CROSS_TC} --disable-static --enable-shared --without-magick-plus-plus --disable-openmp --disable-deprecated --disable-installed --disable-hdri --disable-opencl --disable-largefile --with-threads --without-modules --with-quantum-depth=8 --without-perl --without-bzlib --without-x --with-zlib --without-autotrace --without-dps --without-djvu --without-fftw --without-fpx --without-fontconfig --with-freetype --without-gslib --without-gvc --without-jbig --with-jpeg --without-openjp2 --without-lcms --without-lcms --without-lqr --without-lzma --without-openexr --without-pango --with-png --without-rsvg --without-tiff --without-webp --without-wmf --without-xml
make ${JOBSFLAGS} V=1
make install

cp build/lib/pkgconfig/* /usr/lib/arm-linux-gnueabi/pkgconfig/