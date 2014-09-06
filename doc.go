/*
Package duplo provides tools to efficiently query large sets of images for
visual duplicates. The technique is based on the paper "Fast Multiresolution
Image Querying" by Charles E. Jacobs, Adam Finkelstein, and David H. Salesin,
with a few modifications and additions, such as the addition of a width to
height ratio, the dHash metric by Dr. Neal Krawetz as well as some
histogram-based metrics.

Quering the data structure will return a list of potential matches, sorted by
the score described in the main paper. The user can make searching for
duplicates stricter, however, by filtering based on the additional metrics.
*/
package duplo
