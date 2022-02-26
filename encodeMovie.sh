ffmpeg -f image2 -framerate 30 -pattern_type glob -i "*?png" -vcodec libx264 -crf 18 -pix_fmt yuv420p output.mp4

