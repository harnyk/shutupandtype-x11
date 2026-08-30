package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework AppKit -framework Foundation
#include <stdlib.h>
#import <AppKit/AppKit.h>

static void setPasteboardUTF8(const char *utf8) {
	@autoreleasepool {
		NSString *s = [NSString stringWithUTF8String:utf8];
		if (s == nil) {
			return;
		}
		NSPasteboard *pb = [NSPasteboard generalPasteboard];
		[pb clearContents];
		[pb setString:s forType:NSPasteboardTypeString];
	}
}
*/
import "C"
import "unsafe"

func toClipboard(text string) error {
	c := C.CString(text)
	defer C.free(unsafe.Pointer(c))
	C.setPasteboardUTF8(c)
	return nil
}
