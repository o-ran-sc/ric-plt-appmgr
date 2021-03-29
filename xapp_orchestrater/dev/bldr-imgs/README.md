# Base builder images for O-RAN-SC

This area contains Dockerfiles for images suitable for use as the
first stage in a multi-stage Docker build.  These images have large
build and compile tools like C, C++, Golang, cmake, ninja, etc.  Using
these base images reduces the time needed to build Docker images with
project features. The images are published to the O-RAN-SC staging
registry at the Linux Foundation: nexus3.o-ran-sc.org:10004
