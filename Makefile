cmds = glyph_maker glyph_tester load_img example

export build = $(abspath ./build)

build_windows_amd64=$(abspath ./build/windows_amd64)
build_darwin_amd64=$(abspath ./build/darwin_amd64)

generator=gen_element

dsts=$(addprefix $(build)/,$(cmds))
dsts_windows_amd64=$(addprefix $(build_windows_amd64)/,example)
dsts_darwin_amd64=$(addprefix $(build_darwin_amd64)/,example)

pkg=$(shell find . -name \*.go)

#build_flags=-ldflags="-w -extldflags=-static" 

all: $(generator) $(dsts)

example: ./build/example

ms: $(dsts_windows_amd64)

mac: $(dsts_darwin_amd64)

.SECONDEXPANSION:

$(generator): $$(shell find ./cmd/gen_element -name \*.go) | $(build)
	cd $(dir $<); \
	go build -o $(abspath $@)

$(dsts): $$(shell find ./cmd/$$(notdir $$@) -name \*.go) $(pkg) | $(build)
	export CGO_ENABLE=1; \
	go generate; \
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
