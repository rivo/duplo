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
The following thresholds for those metrics have been found by experimentation
but you may need to adjust them based on your specific imagery.

  // Any two image ratios which differ by more than this amount are considered
  // no duplicates.
  ratioThreshold = 0.1

  // Any hamming distances between two dHash bit vectors greater than this
  // number are ignored during matching.
  dHashThreshold = 6

  // Any hamming distances between two histogram bit vectors greater than this
  // number are ignored during matching.
  histogramThreshold = 7

  // Any absolute difference between two histogram maxima greater than this
  // number are ignored during matching.
  histoMaxThreshold = 0.13

  // Any matching score higher than this value will disqualify an image as a
  // duplicate.
  scoreThreshold = 1200.0

*/
package duplo
