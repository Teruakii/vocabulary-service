package auth

import (
	"sync"
	"time"
)

// loginLimiter blocks password guessing: after maxFails wrong passwords for
// the same email from the same IP within window, that pair is locked out for
// lockout. Keying on IP+email means one attacker cannot lock a victim out from
// everywhere, and one user's typos do not affect anyone else.
type loginLimiter struct {
	mu       sync.Mutex
	entries  map[string]*loginAttempts
	maxFails int
	window   time.Duration
	lockout  time.Duration
	now      func() time.Time
}

type loginAttempts struct {
	fails       int
	firstFail   time.Time
	lockedUntil time.Time
}

func newLoginLimiter(maxFails int, window, lockout time.Duration) *loginLimiter {
	return &loginLimiter{
		entries:  map[string]*loginAttempts{},
		maxFails: maxFails,
		window:   window,
		lockout:  lockout,
		now:      time.Now,
	}
}

// retryAfter returns how long key must wait before trying again (0 = allowed).
func (l *loginLimiter) retryAfter(key string) time.Duration {
	l.mu.Lock()
	defer l.mu.Unlock()
	a, ok := l.entries[key]
	if !ok {
		return 0
	}
	if wait := a.lockedUntil.Sub(l.now()); wait > 0 {
		return wait
	}
	return 0
}

func (l *loginLimiter) fail(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	l.sweep(now)

	a, ok := l.entries[key]
	if !ok || now.Sub(a.firstFail) > l.window {
		a = &loginAttempts{firstFail: now}
		l.entries[key] = a
	}
	a.fails++
	if a.fails >= l.maxFails {
		a.lockedUntil = now.Add(l.lockout)
		a.fails = 0
		a.firstFail = now
	}
}

func (l *loginLimiter) succeed(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.entries, key)
}

// sweep drops stale entries so memory stays bounded; caller holds mu.
func (l *loginLimiter) sweep(now time.Time) {
	if len(l.entries) < 10_000 {
		return
	}
	for k, a := range l.entries {
		if now.After(a.lockedUntil) && now.Sub(a.firstFail) > l.window {
			delete(l.entries, k)
		}
	}
}
