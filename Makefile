ASSETS_PATH=./assets

BIN_SOURCE_PATH=.
BIN_SOURCE_FILE=finance-planner

BIN_TARGET_PATH=${HOME}/.local/bin/
BIN_TARGET_FILE=finance-planner

DESKTOP_SOURCE_PATH=${ASSETS_PATH}
DESKTOP_SOURCE_FILE=Finance Planner.desktop

DESKTOP_TARGET_PATH=${HOME}/.local/share/applications
DESKTOP_TARGET_FILE=Finance Planner.desktop

ICON_SOURCE_PATH=${ASSETS_PATH}
ICON_SOURCE_FILE=icon-256.png

ICON_TARGET_PATH=${HOME}/.local/share/icons/hicolor/256x256/apps/
ICON_TARGET_FILE=finance-planner-256.png

build:
	go get -v
	go build -v
	echo "Done building. Suggest running 'make install' to create a desktop application entry."

install:
	echo "If this fails, then you need to run 'make build' first." && test -f "${BIN_SOURCE_PATH}/${BIN_SOURCE_FILE}" || exit 1
	mkdir -p "${BIN_TARGET_PATH}"
	mkdir -p "${DESKTOP_TARGET_PATH}"
	mkdir -p "${ICON_TARGET_PATH}"
	cp "${BIN_SOURCE_FILE}" "${BIN_TARGET_PATH}/${BIN_TARGET_FILE}"
	cp "${DESKTOP_SOURCE_PATH}/${DESKTOP_SOURCE_FILE}" "${DESKTOP_TARGET_PATH}/${DESKTOP_TARGET_FILE}"
	cp "${ICON_SOURCE_PATH}/${ICON_SOURCE_FILE}" "${ICON_TARGET_PATH}/${ICON_TARGET_FILE}"
	sed -i "s|@@HOME@@|${HOME}|g" "${DESKTOP_TARGET_PATH}/${DESKTOP_TARGET_FILE}"
	echo "Done installing. Please make sure your PATH variable contains ~/.local/bin"

uninstall:
	-rm "${DESKTOP_TARGET_PATH}/${DESKTOP_TARGET_FILE}"
	-rm "${ICON_TARGET_PATH}/${ICON_TARGET_FILE}"
	-rm "${BIN_TARGET_PATH}"
	echo "Done uninstalling."
