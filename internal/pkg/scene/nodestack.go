package scene

type SceneNodeStack []*SceneNode

// IsEmpty checks if stack is empty
func (stack *SceneNodeStack) IsEmpty() bool {
	return len(*stack) == 0
}

// Push pushes a new scene node to the top of the stack.
func (stack *SceneNodeStack) Push(sceneNode *SceneNode) {
	*stack = append(*stack, sceneNode)
}

// PushAll pushes all new scene node to the top of the stack.
func (stack *SceneNodeStack) PushAll(sceneNodes []*SceneNode) {
	*stack = append(*stack, sceneNodes...)
}

// Pop removes and return top element of stack.
// Return false if stack is empty.
func (stack *SceneNodeStack) Pop() (*SceneNode, bool) {
	if stack.IsEmpty() {
		return nil, false
	}

	topIndex := len(*stack) - 1     // Get the topIndex of the top most element.
	sceneNode := (*stack)[topIndex] // Index into the slice and obtain the element.
	(*stack)[topIndex] = nil        // Erase top element entry (write zero value)
	*stack = (*stack)[:topIndex]    // Remove it from the stack by slicing it off.

	return sceneNode, true
}
