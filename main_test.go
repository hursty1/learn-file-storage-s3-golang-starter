package main

import (
	"fmt"
	"os"
	"testing"
)


func TestCommandRun(t *testing.T) {
	ratio, err := getVideoAspectRatio("samples/boots-video-vertical.mp4")
	if err != nil {
		t.Errorf("Failed to execute %s", err)
	}
	if ratio != "9:16" {
		t.Errorf("Expected 9:16 and got %s", ratio)
	}

	ratio, err = getVideoAspectRatio("samples/boots-video-horizontal.mp4")
	if err != nil {
		t.Errorf("Failed to execute %s", err)
	}
	if ratio != "16:9" {
		t.Errorf("Expected 16:9 and got %s", ratio)
	}
}

func TestCommandFFmpeg(t *testing.T){
	cases := [] struct {
		filepath string
		expecting string
	} {
		{
			filepath: "samples/boots-video-vertical.mp4",
			expecting: "samples/boots-video-vertical.mp4" + ".processing",
		},
		{
			filepath: "samples/boots-video-vertical.mp4",
			expecting: "samples/boots-video-vertical.mp4" + ".processing",
		},
	}
	for _, c := range cases {
		
		output, err := processVideoForFastStart(c.filepath)
		if err != nil {
			t.Errorf("Failed to execute %s", err)
		}
		if output != c.expecting {
			t.Errorf("Output filepath does not match expected")
		}
		fmt.Printf("Output is: %s\n", output)
		err = os.Remove(output)
		if err != nil {
			fmt.Println("Failed to clean up test")
		}

	}
}