package gltf

type AnimationChannelTargetPath string

const (
	AnimationChannelTargetPath_TRANSLATION AnimationChannelTargetPath = "translation"
	AnimationChannelTargetPath_ROTATION    AnimationChannelTargetPath = "rotation"
	AnimationChannelTargetPath_SCALE       AnimationChannelTargetPath = "scale"
	AnimationChannelTargetPath_WEIGHTS     AnimationChannelTargetPath = "weights"
)

// The descriptor of the animated property.
// https://github.com/KhronosGroup/glTF/blob/main/specification/2.0/schema/animation.channel.target.schema.json
type AnimationChannelTarget struct {
	Property
	Node *GltfId `json:"node,omitempty"` // The index of the node to animate. When undefined, the animated object **MAY** be defined by an extension.

	// The name of the node's TRS property to animate, or the "weights" of the
	// Morph Targets it instantiates. For the "translation" property, the
	// values that are provided by the sampler are the translation along the
	// X, Y, and Z axes. For the `\"rotation\"` property, the values are a
	// quaternion in the order (x, y, z, w), where w is the scalar. For the
	// "scale" property, the values are the scaling factors along the
	// X, Y, and Z axes.
	Path AnimationChannelTargetPath `json:"path"`
}

// An animation channel combines an animation sampler with a target property being animated.
// https://github.com/KhronosGroup/glTF/blob/main/specification/2.0/schema/animation.channel.schema.json
type AnimationChannel struct {
	Property
	Sampler GltfId                 `json:"sampler"` // The index of a sampler in this animation used to compute the value for the target, e.g., a node's translation, rotation, or scale (TRS).
	Target  AnimationChannelTarget `json:"target"`  // The descriptor of the animated property.
}

type AnimationSamplerInterpolation string

const (
	AnimationSamplerInterpolation_LINEAR      AnimationSamplerInterpolation = "LINEAR"      // The animated values are linearly interpolated between keyframes. When targeting a rotation, spherical linear interpolation (slerp) **SHOULD** be used to interpolate quaternions. The number of output elements **MUST** equal the number of input elements.
	AnimationSamplerInterpolation_STEP        AnimationSamplerInterpolation = "STEP"        // The animated values remain constant to the output of the first keyframe, until the next keyframe. The number of output elements **MUST** equal the number of input elements.
	AnimationSamplerInterpolation_CUBICSPLINE AnimationSamplerInterpolation = "CUBICSPLINE" // The animation's interpolation is computed using a cubic spline with specified tangents. The number of output elements **MUST** equal three times the number of input elements. For each input element, the output stores three elements, an in-tangent, a spline vertex, and an out-tangent. There **MUST** be at least two keyframes when using this interpolation.
)

// An animation sampler combines timestamps with a sequence of output values and defines an interpolation algorithm.
// https://github.com/KhronosGroup/glTF/blob/main/specification/2.0/schema/animation.sampler.schema.json
type AnimationSampler struct {
	Property
	Input         GltfId                        `json:"input"`                   // The index of an accessor containing keyframe timestamps. The accessor **MUST** be of scalar type with floating-point components. The values represent time in seconds with `time[0] >= 0.0`, and strictly increasing values, i.e., `time[n + 1] > time[n]`.
	Output        GltfId                        `json:"output"`                  // The index of an accessor, containing keyframe output values.
	Interpolation AnimationSamplerInterpolation `json:"interpolation,omitempty"` // The index of an accessor, containing keyframe output values.
}

// A keyframe animation.
// https://github.com/KhronosGroup/glTF/blob/main/specification/2.0/schema/animation.schema.json
type Animation struct {
	ChildOfRootProperty
	Channels []AnimationChannel `json:"channels"` // An array of animation channels. An animation channel combines an animation sampler with a target property being animated. Different channels of the same animation **MUST NOT** have the same targets.
	Samplers []AnimationSampler `json:"samplers"` // An array of animation samplers. An animation sampler combines timestamps with a sequence of output values and defines an interpolation algorithm.
}
