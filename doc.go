/*
Package duplo provides tools to efficiently query large sets of images for
visual duplicates. The technique is based on the paper "Fast Multiresolution
Image Querying" by Charles E. Jacobs, Adam Finkelstein, and David H. Salesin,
with a few modifications and additions. The main goal is to find actual
duplicates. The search is therefore a bit more strict than if we had to find
only similar images.

Specifically, we also consider the ratio of image width to image height,
excluding dissimilar ratios.
*/
package duplo
