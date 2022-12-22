# Encodes a movie from still images.
# Just run this script in the directory with all the animation still images.
#
# Images should have an index number in the filename i.e. like "my_animation_00042.png" to be picked up
# by the encoder in the right order.

ffmpeg -hide_banner -loglevel error -f image2 -framerate 25 -pattern_type glob -i "*?png" -vcodec libx264 -crf 18 -pix_fmt yuv420p output.mp4


