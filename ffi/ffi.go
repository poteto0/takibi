// Package ffi converts values between Go and JavaScript.
//
// Every function needs syscall/js, so the package is empty on native builds:
// import it only from files carrying the wasm build tag.
package ffi
