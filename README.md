Go Globals
==========

This module is a bunch of stuff I always call in my main() function,
setting up logging and my config system and a signal handler to reload
the config and metric reporter, that sort of thing.

I was going to call it "go-main" or something, but it really is managing
the per-process globals so I went with that name.

