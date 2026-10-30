package ops

import "testing"

func TestMultiListReadIndexesOnce(t *testing.T) {
	f, _ := chatFixture(t)
	original := f.Index
	reads := 0
	f.Index = func(path string) (map[string]string, error) { reads++; return original(path) }
	_, groups, err := f.ChatLists()
	must(t, err)
	count := 0
	for _, chats := range groups {
		count += len(chats)
	}
	if reads != 1 || count != 3 {
		t.Fatalf("index reads=%d, chats=%d", reads, count)
	}
	_, _, err = f.ChatLists()
	must(t, err)
	if reads != 2 {
		t.Fatal("a later read reused a stale transcript index")
	}
}

func TestDoctorIndexesOnce(t *testing.T) {
	f, _ := chatFixture(t)
	original := f.Index
	reads := 0
	f.Index = func(path string) (map[string]string, error) { reads++; return original(path) }
	f.Doctor()
	if reads != 1 {
		t.Fatalf("doctor indexed transcripts %d times", reads)
	}
}

func TestChatListsReadsFreshCards(t *testing.T) {
	f, now := chatFixture(t)
	_, first, err := f.ChatLists()
	must(t, err)
	f.card(work, "c1", s1, "Updated title", now, "")
	_, second, err := f.ChatLists()
	must(t, err)
	found := false
	for _, group := range second {
		for _, c := range group {
			if c.ID == "local_c1" {
				found = c.Title == "Updated title"
			}
		}
	}
	if !found || len(first) != len(second) {
		t.Fatal("chat snapshot retained stale cards")
	}
}

func TestEmptyChatListsSkipTranscriptIndex(t *testing.T) {
	f := newFixture(t)
	f.Index = func(string) (map[string]string, error) { t.Fatal("empty lists indexed transcripts"); return nil, nil }
	lists, groups, err := f.ChatLists()
	must(t, err)
	if len(lists) != 0 || len(groups) != 0 {
		t.Fatal("empty setup returned chats")
	}
}
