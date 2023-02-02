// Copyright © 2021-2023 The Gomon Project.

package gocore

/*
#cgo CFLAGS: -x objective-c -std=gnu11 -fobjc-arc
#cgo LDFLAGS: -framework Foundation -framework AppKit
#import <Foundation/Foundation.h>
#import <AppKit/AppKit.h>

// The following code invoked by C.run() is to capture dynamically the system appearance.
// Simply querying the NSApplication effective appearance is insufficient if no view defined.

// darkmode sets flag in Go code, it is defined in core_darwin.go.
extern void darkmode(bool);

@interface MyView : NSView
@end

@implementation MyView
- (id)initWithFrame:(CGRect)frame
{
    self = [super initWithFrame:frame];
    return self;
}
- (void)viewDidChangeEffectiveAppearance {
	NSString *name = [[self effectiveAppearance] name];
	NSLog(@"Changed Appearance is %@", name);
	name = [[self effectiveAppearance]
		bestMatchFromAppearancesWithNames: @[NSAppearanceNameAqua, NSAppearanceNameDarkAqua]];
	darkmode((name == NSAppearanceNameDarkAqua) ? TRUE : FALSE);
}
@end

@interface MyAppDelegate : NSObject<NSApplicationDelegate>
@end

@implementation MyAppDelegate
- (void)applicationDidFinishLaunching:(NSNotification*)note {
	NSLog(@"Finish launch notification is %@", [note description]);
}
- (NSApplicationTerminateReply)applicationShouldTerminate:(NSApplication *)app {
	NSLog(@"Application should terminate %@", [app description]);
	return NSTerminateNow;
}
@end

static void
run() {
	[NSApplication sharedApplication]; // initialize the application
	[NSApp setDelegate: [[MyAppDelegate alloc] init]];

	NSRect rect = NSMakeRect(100.0, 100.0, 100.0, 100.0);
	NSWindow *window = [[NSWindow alloc]
		initWithContentRect:rect
           	      styleMask:NSWindowStyleMaskTitled|NSWindowStyleMaskClosable
               	    backing:NSBackingStoreBuffered
                   	  defer:NO
	];

	rect = [[window contentView] frame];
	MyView *view = [[MyView alloc] initWithFrame: rect];
	[window setContentView: view];
	[NSApp hide:nil];

	[NSApp run];
}

static void
terminate() {
	[NSApp terminate: nil];
}
*/
import "C"

// osEnvironment starts the native application environment run loop.
// Note that a native application environment runs on the main thread.
// Therefore, launch the gomon command Main() in a go routine.
// Unfortunately, attempting to initialize the macOS NSApplication can
// occasionally result in
//   Terminating app due to uncaught exception 'NSInternalInconsistencyException',
//   reason: 'NSWindow drag regions should only be invalidated on the Main Thread!'
// Therefore, abandoning this attempt at NSApp integration.

// func osEnvironment(ctx context.Context) {
// 	defer func() {
// 		if r := recover(); r != nil {
// 			buf := make([]byte, 4096)
// 			n := runtime.Stack(buf, false)
// 			buf = buf[:n]
// 			LogError(fmt.Errorf("osEnvironment panicked, %v\n%s", r, buf))
// 		}
// 	}()

// 	runtime.LockOSThread() // tie this goroutine to the main OS thread
// 	defer runtime.UnlockOSThread()

// 	go func() {
// 		<-ctx.Done()
// 		C.terminate()
// 	}()
// 	C.run()
// }
