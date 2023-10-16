PREFIX     = /usr/local
MANPREFIX  = ${PREFIX}/share/man
DESKPREFIX = ${PREFIX}/share/applications

all: prep lf

prep: go install
	env CGO_ENABLED=0 go install -ldflags="-s -w" github.com/gokcehan/lf@latest

lf: lf
	go build

install:
	mkdir -p ${DESTDIR}${PREFIX}/bin
	cp -f lf ${DESTDIR}${PREFIX}/bin
	chmod 755 ${DESTDIR}${PREFIX}/bin/lf
	mkdir -p ${DESTDIR}${MANPREFIX}/man1
	cp -f lf.1 ${DESTDIR}${MANPREFIX}/man1/lf.1
	chmod 644 ${DESTDIR}${MANPREFIX}/man1/lf.1
	mkdir -p ${DESTDIR}${DESKPREFIX}
	cp -f lf.desktop ${DESTDIR}${DESKPREFIX}/lf.desktop

uninstall:
	rm -f ${DESTDIR}${PREFIX}/bin/lf
	rm -f ${DESTDIR}${MANPREFIX}/man1/lf.1
	rm -f ${DESTDIR}${DESKPREFIX}/lf.desktop


.PHONY: all prep install uninstall
