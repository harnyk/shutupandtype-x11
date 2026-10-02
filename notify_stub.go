//go:build !darwin && !linux

package main

func platformShowNotification(title, body string) {}
