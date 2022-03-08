RENDER_RESULT_DIR=sphere_circle_rotation
SCENE_BIN=animation_sphere_circle_rotation
SCENE_DEFINITION=sphere_circle_rotation.animation.json

clear

# make build_sphere_rotation
make build_all

# Run animation file creation program
./bin/$SCENE_BIN

# Wipe/clear output directory
rm -fR ./rendered/$RENDER_RESULT_DIR

# Render animation/scene
./bin/pathtracer scene/$SCENE_DEFINITION

# Encode movie from rendered images
cd ./rendered/$RENDER_RESULT_DIR
../../encodeMovie.sh
cd -
