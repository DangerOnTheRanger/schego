Schego - A Scheme implementation written in Go
==============================================


**shay**-go: noun

definition: yet another Scheme implementation that probably didn't need to exist,
but does anyway.


What?
-----

The title is pretty descriptive. But it's missing some details. In addition to
being a (soon-to-be) fully R7RS-compatible Scheme implementation, Schego will also
include a stack-based, language-agnostic VM, suitable for implementing other
languages instead of Scheme.

At some point in the future, Schego's VM will also include a JIT targeting
the x86-64/amd64 platform (and perhaps some ARM devices), in addition to
being able to be embedded into any Go application to provide Scheme scripting support, or just
a generic VM to build on top of. Why would you want to do that when LLVM exists? I'm not entirely sure. But you could.

To summarize Schego's (planned) feature set in a quick bullet point list:

* Complete R7RS Scheme implementation
* R7RS SLib (https://github.com/petercrlane/r7rs-libs) compatibility
* Hyrbrid JIT backend targeting x86-64/amd64 and maybe possibly some ARM devices down the road
* libschego build target allowing embedding into almost any Go application
* STM (Software Transactional Memory) multitasking model, possibly the first of its kind
  in Scheme


Why?
----

To familiarize myself with Go, primarily. There was, at least at the time I write this,
very little practical need for a serious Scheme implementation in Go - many good,
established Scheme implementations are already out there, with some of them even
having basically the entire feature set I initially planned for Schego.

I do intend to implement Schego with a watchful eye towards readability, so
hopefully the codebase will be able to serve as a reasonable frame of reference
towards anyone trying to implement something similar. I'll also cover Schego's
progress on my blog, which will hopefully be educational and at least slightly entertaining.

That, and I needed an excuse to do something with Scheme.
