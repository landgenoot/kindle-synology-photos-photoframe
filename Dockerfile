FROM debian:bullseye

RUN apt-get update && apt-get install -y \
    curl \
    build-essential \
    automake \
    autogen \
    bash \
    build-essential \
    bc \
    bzip2 \
    ca-certificates \
    curl \
    file \
    git \
    gzip \
    make \
    ncurses-dev \
    pkg-config \
    libtool \
    python \
    rsync \
    sed \
    bison \
    flex \
    tar \
    vim \
    wget \
    runit \
    xz-utils \
    golang 

RUN dpkg --add-architecture armel
RUN apt-get update && apt-get install -y crossbuild-essential-armel \
    libmagickwand-dev:armel \
    libmagickcore-dev:armel \
    libfftw3-dev:armel

# The cross-compiling emulator
RUN apt-get update && apt-get install -y \
    qemu-user \
    qemu-user-static

ADD build-image-magick.sh /home

RUN chmod +x /home/build-image-magick.sh && /home/build-image-magick.sh
