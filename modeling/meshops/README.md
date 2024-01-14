# Mesh Operations

A package serving for the implementation of common mesh operations.

## Implemented Operations

## Center Attribute

Calculates the AABB for the attribute specified (most commonly position, but could be used for anything like UV Coordinates) and offsets all vertice data by the center of the AABB.

### Flat Normals

We set each vertices normal to be equal to the face's normal. If the vertice is used for multiple faces, a face is arbitrarily chosen. If you want to avoid this behavior, you should run [Unweld](#unweld) first 

### Flip Winding

### Laplacian Smoothing

### Normalize Attribute

### Remove Null Faces

### Remove Unreferenced Vertices

### Rotate Attribute

### Scale Attribute

### Smooth Normals

### Translate Attribute

### Unweld

### Vertex Color Space