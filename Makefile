GOMOBILE=gomobile
GOBIND=$(GOMOBILE) bind
BUILDDIR=$(shell pwd)/build
IOS_ARTIFACT=$(BUILDDIR)/Watermob.xcframework
ANDROID_ARTIFACT=$(BUILDDIR)/watermob.aar
IOS_TARGET=ios,iossimulator
ANDROID_TARGET=android
ANDROID_API=27
# LDFLAGS='-s -w -X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=warn'
LDFLAGS='-s -w'
IMPORT_PATH=github.com/gaukas/watermob

BUILD_IOS="$(GOBIND) -a -ldflags $(LDFLAGS) -target=$(IOS_TARGET) -o $(IOS_ARTIFACT)"
BUILD_ANDROID="$(GOBIND) -a -ldflags $(LDFLAGS) -target=$(ANDROID_TARGET) -androidapi $(ANDROID_API) -tags=gomobile -o $(ANDROID_ARTIFACT)"

all: ios android

ios:
	mkdir -p $(BUILDDIR)
	eval $(BUILD_IOS)

android:
	rm -rf $(BUILDDIR) 2>/dev/null
	mkdir -p $(BUILDDIR)
	eval $(BUILD_ANDROID)

clean:
	rm -rf $(BUILDDIR)