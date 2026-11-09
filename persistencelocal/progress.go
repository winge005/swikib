package persistencelocal

import "sync"

type ProgressSnapshot struct {
	Action  string `json:"action"`
	Running bool   `json:"running"`
	Paused  bool   `json:"paused"`
	Percent int    `json:"percent"`
	Message string `json:"message"`
	Error   string `json:"error"`
}

type jobProgress struct {
	mu             sync.RWMutex
	action         string
	running        bool
	paused         bool
	pauseRequested bool
	percent        int
	message        string
	errMsg         string
}

var currentJob = &jobProgress{}

func TryStartProgress(action string) bool {
	currentJob.mu.Lock()
	defer currentJob.mu.Unlock()

	if currentJob.running {
		return false
	}

	currentJob.action = action
	currentJob.running = true
	currentJob.paused = false
	currentJob.pauseRequested = false
	currentJob.percent = 0
	currentJob.message = "Bezig..."
	currentJob.errMsg = ""

	return true
}

// RequestPause asks the running analyse job to stop at the next safe point.
// It returns false when there is no analyse job to pause.
func RequestPause() bool {
	currentJob.mu.Lock()
	defer currentJob.mu.Unlock()

	if !currentJob.running || currentJob.action != "analyse" {
		return false
	}

	currentJob.pauseRequested = true

	return true
}

func PauseRequested() bool {
	currentJob.mu.RLock()
	defer currentJob.mu.RUnlock()

	return currentJob.pauseRequested
}

func FinishPausedProgress() {
	currentJob.mu.Lock()
	defer currentJob.mu.Unlock()

	currentJob.running = false
	currentJob.paused = true
	currentJob.pauseRequested = false
	currentJob.message = "Gepauzeerd"
}

func setProgress(percent int, message string) {
	currentJob.mu.Lock()
	defer currentJob.mu.Unlock()

	if !currentJob.running {
		return
	}

	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}
	if percent < currentJob.percent {
		percent = currentJob.percent
	}

	currentJob.percent = percent
	if message != "" {
		currentJob.message = message
	}
}

func FinishProgress(err error) {
	currentJob.mu.Lock()
	defer currentJob.mu.Unlock()

	if err != nil {
		currentJob.errMsg = err.Error()
		currentJob.message = err.Error()
	} else if currentJob.percent < 100 {
		currentJob.percent = 100
		currentJob.message = "Klaar"
	}

	currentJob.running = false
	currentJob.paused = false
	currentJob.pauseRequested = false
}

func CurrentProgress() ProgressSnapshot {
	currentJob.mu.RLock()
	defer currentJob.mu.RUnlock()

	return ProgressSnapshot{
		Action:  currentJob.action,
		Running: currentJob.running,
		Paused:  currentJob.paused,
		Percent: currentJob.percent,
		Message: currentJob.message,
		Error:   currentJob.errMsg,
	}
}

func stepPercent(base, span, done, total int) int {
	if total <= 0 {
		return base
	}

	percent := base + done*span/total
	if percent > base+span {
		return base + span
	}

	return percent
}
