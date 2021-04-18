package main

import "math"

func idf(docFreq, docCount float64) float64 {
	return math.Log(1 + (docCount-docFreq+0.5)/(docFreq+0.5))
}

func tf(termFreq, fieldLen, avgFieldLen float64) float64 {
	return (termFreq * 2.2) / (termFreq + 1.2*(0.25+0.75*(fieldLen/avgFieldLen)))
}
