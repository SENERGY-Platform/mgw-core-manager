/*
 * Copyright 2024 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package log_hdl

import (
	"bytes"
	"context"
	"io"
)

var sep = []byte{'\n'}

func seek(ctx context.Context, rs io.ReadSeeker, numLines int, bufSize int64) (int64, error) {
	if numLines <= 0 {
		return 0, nil
	}
	pos, err := rs.Seek(0, io.SeekEnd)
	if err != nil {
		return 0, err
	}
	if pos <= 0 {
		return 0, nil
	}
	if pos-bufSize < 0 {
		bufSize = pos
	}
	var buffer []byte
	var bufLineCount int
	lineCount := 0
	for pos > 0 {
		if ctx.Err() != nil {
			return 0, ctx.Err()
		}
		pos = pos - bufSize
		if pos < 0 {
			bufSize = bufSize + pos
			pos = 0
		}
		_, err = rs.Seek(pos, io.SeekStart)
		if err != nil {
			return 0, err
		}
		bufLineCount, buffer, err = countLines(rs, bufSize)
		if err != nil {
			return 0, err
		}
		lineCount += bufLineCount
		if lineCount > numLines {
			break
		}
	}
	numLinesToRead := numLines - (lineCount - bufLineCount)
	switch {
	case lineCount == 0:
		pos = 0
	case lineCount < numLines:
		pos = 0
	case numLinesToRead > 0:
		pos = pos + bufSize
		bufIndex := bytes.LastIndex(buffer, sep)
		for numLinesToRead > 0 {
			if ctx.Err() != nil {
				return 0, ctx.Err()
			}
			bufIndex = bytes.LastIndex(buffer[:bufIndex], sep)
			if bufIndex < 0 {
				bufIndex = 0
				break
			}
			numLinesToRead--
		}
		pos = pos - (bufSize - int64(bufIndex))
		if pos > 0 {
			pos++
		}
	case lineCount-bufLineCount == numLines:
		bufIndex := bytes.LastIndex(buffer, sep)
		pos = pos + int64(bufIndex)
		if pos > 0 {
			pos++
		}
	}
	if pos < 0 {
		pos = 0
	}
	return rs.Seek(pos, io.SeekStart)
}

func countLines(r io.Reader, size int64) (int, []byte, error) {
	buf := make([]byte, size)
	_, err := r.Read(buf)
	if err != nil && err != io.EOF {
		return 0, nil, err
	}
	return bytes.Count(buf, sep), buf, nil
}
