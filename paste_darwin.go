package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework AppKit -framework Foundation -framework CoreGraphics -framework Carbon
#import <CoreGraphics/CoreGraphics.h>
#import <Carbon/Carbon.h>

static void postCommandV(void) {
	CGEventSourceRef src = CGEventSourceCreate(kCGEventSourceStateHIDSystemState);
	CGEventRef down = CGEventCreateKeyboardEvent(src, (CGKeyCode)kVK_ANSI_V, true);
	CGEventRef up = CGEventCreateKeyboardEvent(src, (CGKeyCode)kVK_ANSI_V, false);
	CGEventSetFlags(down, kCGEventFlagMaskCommand);
	CGEventSetFlags(up, kCGEventFlagMaskCommand);
	CGEventPost(kCGHIDEventTap, down);
	CGEventPost(kCGHIDEventTap, up);
	CFRelease(down);
	CFRelease(up);
	CFRelease(src);
}
*/
import "C"

// typeShiftInsert pastes via Cmd+V using CGEvent (needs Accessibility).
func typeShiftInsert() error {
	C.postCommandV()
	return nil
}
