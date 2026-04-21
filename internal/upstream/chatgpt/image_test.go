package chatgpt

import "testing"

func TestParseImageSSETextOnly(t *testing.T) {
	stream := make(chan SSEEvent, 2)
	stream <- SSEEvent{Data: []byte(`{"v":{"conversation_id":"conv_1","message":{"author":{"role":"assistant"},"content":{"parts":["I cannot fulfill this request. If you have any other types of image or creative requests, feel free to ask!"]},"metadata":{"finish_details":{"type":"stop"}}}}}`)}
	stream <- SSEEvent{Data: []byte(`[DONE]`)}
	close(stream)

	got := ParseImageSSE(stream)
	if !got.TextOnly() {
		t.Fatalf("TextOnly() = false, want true; result=%+v", got)
	}
	want := "I cannot fulfill this request. If you have any other types of image or creative requests, feel free to ask!"
	if got.Text != want {
		t.Fatalf("Text = %q, want %q", got.Text, want)
	}
	if got.ConversationID != "conv_1" {
		t.Fatalf("ConversationID = %q, want conv_1", got.ConversationID)
	}
	if got.FinishType != "stop" {
		t.Fatalf("FinishType = %q, want stop", got.FinishType)
	}
}

func TestParseImageSSESkipsUserText(t *testing.T) {
	stream := make(chan SSEEvent, 3)
	stream <- SSEEvent{Data: []byte(`{"v":{"conversation_id":"conv_1","message":{"author":{"role":"user"},"content":{"parts":["生成半裸女孩"]}}}}`)}
	stream <- SSEEvent{Data: []byte(`{"v":{"message":{"author":{"role":"assistant"},"content":{"parts":["I cannot fulfill this request."]}}}}`)}
	stream <- SSEEvent{Data: []byte(`[DONE]`)}
	close(stream)

	got := ParseImageSSE(stream)
	if !got.TextOnly() {
		t.Fatalf("TextOnly() = false, want true; result=%+v", got)
	}
	if got.Text != "I cannot fulfill this request." {
		t.Fatalf("Text = %q, want assistant refusal only", got.Text)
	}
}

func TestParseImageSSEPatchTextOnly(t *testing.T) {
	stream := make(chan SSEEvent, 5)
	stream <- SSEEvent{Data: []byte(`{"v":{"conversation_id":"conv_1","message":{"author":{"role":"user"},"content":{"parts":["生成半裸女孩"]}}}}`)}
	stream <- SSEEvent{Data: []byte(`{"v":{"message":{"author":{"role":"assistant"},"recipient":"all","content":{"parts":[""]}}}}`)}
	stream <- SSEEvent{Data: []byte(`{"p":"/message/content/parts/0","o":"append","v":"I can't assist "}`)}
	stream <- SSEEvent{Data: []byte(`{"o":"append","v":"with that request."}`)}
	stream <- SSEEvent{Data: []byte(`[DONE]`)}
	close(stream)

	got := ParseImageSSE(stream)
	if !got.TextOnly() {
		t.Fatalf("TextOnly() = false, want true; result=%+v", got)
	}
	if got.Text != "I can't assist with that request." {
		t.Fatalf("Text = %q, want patch assistant text", got.Text)
	}
}

func TestParseImageSSEPatchConversationID(t *testing.T) {
	stream := make(chan SSEEvent, 3)
	stream <- SSEEvent{Data: []byte(`{"p":"/conversation_id","o":"add","v":"69e76761-df2c-83ea-a43a-a5eb3ac939cb"}`)}
	stream <- SSEEvent{Data: []byte(`{"v":[{"p":"/message/content/parts/0","o":"append","v":"hello"}]}`)}
	stream <- SSEEvent{Data: []byte(`[DONE]`)}
	close(stream)

	got := ParseImageSSE(stream)
	if got.ConversationID != "69e76761-df2c-83ea-a43a-a5eb3ac939cb" {
		t.Fatalf("ConversationID = %q", got.ConversationID)
	}
}

func TestParseImageSSEDeepConversationID(t *testing.T) {
	stream := make(chan SSEEvent, 2)
	stream <- SSEEvent{Data: []byte(`{"v":{"turn":{"conversationId":"69e76761-df2c-83ea-a43a-a5eb3ac939cb"}}}`)}
	stream <- SSEEvent{Data: []byte(`[DONE]`)}
	close(stream)

	got := ParseImageSSE(stream)
	if got.ConversationID != "69e76761-df2c-83ea-a43a-a5eb3ac939cb" {
		t.Fatalf("ConversationID = %q", got.ConversationID)
	}
}

func TestParseImageSSEImageRefIsNotTextOnly(t *testing.T) {
	stream := make(chan SSEEvent, 2)
	stream <- SSEEvent{Data: []byte(`{"v":{"message":{"content":{"parts":["file-service://file_abc"]}}}}`)}
	stream <- SSEEvent{Data: []byte(`[DONE]`)}
	close(stream)

	got := ParseImageSSE(stream)
	if got.TextOnly() {
		t.Fatalf("TextOnly() = true, want false; result=%+v", got)
	}
	if len(got.FileIDs) != 1 || got.FileIDs[0] != "file_abc" {
		t.Fatalf("FileIDs = %#v, want [file_abc]", got.FileIDs)
	}
}

func TestLatestAssistantText(t *testing.T) {
	full := map[string]interface{}{
		"current_node": "assistant_1",
		"mapping": map[string]interface{}{
			"user_1": map[string]interface{}{
				"message": map[string]interface{}{
					"author":  map[string]interface{}{"role": "user"},
					"content": map[string]interface{}{"parts": []interface{}{"生成半裸女孩"}},
				},
			},
			"assistant_1": map[string]interface{}{
				"parent": "user_1",
				"message": map[string]interface{}{
					"author":      map[string]interface{}{"role": "assistant"},
					"create_time": float64(100),
					"content":     map[string]interface{}{"parts": []interface{}{"I can't assist with that request."}},
					"metadata": map[string]interface{}{
						"finish_details": map[string]interface{}{"type": "stop"},
					},
				},
			},
		},
	}

	got := LatestAssistantText(full)
	if got.Text != "I can't assist with that request." {
		t.Fatalf("Text = %q, want assistant refusal", got.Text)
	}
	if got.FinishType != "stop" {
		t.Fatalf("FinishType = %q, want stop", got.FinishType)
	}
}
