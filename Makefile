cmds = glyph_maker glyph_tester main 

export build = $(abspath ./build)

build_windows_amd64=$(abspath ./build/windows_amd64)
build_darwin_amd64=$(abspath ./build/darwin_amd64)

dsts=$(addprefix $(build)/,$(cmds))
dsts_windows_amd64=$(addprefix $(build_windows_amd64)/,main)
dsts_darwin_amd64=$(addprefix $(build_darwin_amd64)/,main)

pkg=$(shell find . -name \*.go)

#build_flags=-ldflags="-w -extldflags=-static" 

all: $(dsts)

ms: $(dsts_windows_amd64)

mac: $(dsts_darwin_amd64)

.SECONDEXPANSION:

$(dsts): $$(shell find ./cmd/$$(notdir $$@) -name \*.go) $(pkg) | $(build)
	export CGO_ENABLE=1; \
	cd $(dir $<); \
	go build -o $(abspath $@) $(build_flags)

#export GOGCCFLAGS="-mwindows"; \
#export CGO_CFLAGS="-mwindows"; \

#ms_flags=-ldflags -H=windowsgui

$(dsts_windows_amd64): $$(shell find ./cmd/$$(notdir $$@) -name \*.go) $(pkg) | $(build_windows_amd64)
	export CC=x86_64-w64-mingw32-gcc; \
	export CGO_ENABLED=1; \
	export GOOS=windows; \
	export GOARCH=amd64; \
	cd $(dir $<); \
	go build $(ms_flags) -o $(addsuffix .exe,$(abspath $@))

$(dsts_darwin_amd64): $$(shell find ./cmd/$$(notdir $$@) -name \*.go) $(pkg) | $(build_darwin_amd64)
	export CGO_ENABLED=1; \
	export GOOS=darwin; \
	export GOARCH=amd64; \
	cd $(dir $<); \
	go build -o $(abspath $@) $(build_flags)

$(build) $(build_windows_amd64) $(build_darwin_amd64):
	mkdir -p $@

usb: $(build_windows_amd64)
	sudo mount /dev/sdc1 /media/usb
	sudo cp -t /media/usb $(build_windows_amd64)/*
	sudo umount /media/usb
