NAME =			ssh2docksal
VERSION =		latest
VERSION_ALIASES =
TITLE =			ssh2docksal
DESCRIPTION =		SSH proxy to Docker containers
SOURCE_URL =		https://github.com/andock/ssh2docksal
VENDOR_URL =		https://github.com/andock/ssh2docksal

IMAGE_VOLUME_SIZE =     50G
IMAGE_BOOTSCRIPT =      docker
IMAGE_NAME =            ssh2docksal


## Image tools  (https://github.com/scaleway/image-tools)
all:	docker-rules.mk
docker-rules.mk:
	wget -qO - https://j.mp/scw-builder | bash
-include docker-rules.mk
