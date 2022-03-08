if [ $# -eq 0 ]; then
    echo "No scene name provided as argument."
    exit 1
fi

SCENE_NAME=$1

echo "Building and running scene $SCENE_NAME"

RENDER_RESULT_DIR=$SCENE_NAME
SCENE_BIN=$SCENE_NAME
SCENE_DEFINITION=$SCENE_NAME.animation.json

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
../../encode_movie.sh
mv output.mp4 $SCENE_NAME.mp4
cd -
