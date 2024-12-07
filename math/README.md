# Math

Primitive math utilities to be used by the remaining of the library. Math imports nothing, everything else imports it. There is a slight hierarchy to the sub-packages found in math.

## Sample

Sample serves as a group of definitions for defining a mapping from one numeric value to another. Another term for this is "function" or "mapping" but golang already took both of those keywords so I settled on "sample". Code throughout polyform uses the definitions like `sample.FloatToVec2` to indicate this method maps a float to a Vector 2 value.

There are also some general utility functions for building common mappings between two domains of numbers.

- Compose - Takes an array of sample functions and feeds the output value from  one function into the input value to the next function before finally returning the final value.
    - ComposeFloat
    - ComposeVec2
    - ComposeVec3
- LinearMapping - Remaps some float value between A and B to sample a line in N-Dimensional space linearly.
    - LinearFloatMapping
    - LinearVector2Mapping
    - LinearVector3Mapping
- Trig  
    - Sin - Maps a float value to a sin wave with some specified amplitude and frequency.
    - Cos - Maps a float value to a sin wave with some specified amplitude and frequency.

## Curves

Curves builds off of the sample package. The one restriction it makes is that the input domain is restricted to float values between 0 and 1. This is where more common animation curves reside.

## Noise

Different noise algorithms commonly used in procedural generation.

## SDF

SDF implementations of different geometry primitives, along with common math functions. Basically slowly picking through [Inigo Quilez's Distfunction](https://iquilezles.org/articles/distfunctions/) article as I need them in my different projects.

## TRS

Math around Translation / Rotation / Scale Matrices