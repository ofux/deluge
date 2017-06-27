package object

import "testing"

func TestEnvironment_Get(t *testing.T) {
	t.Run("without outer env", func(t *testing.T) {
		env := NewEnvironment()
		env.store["a"] = &Integer{Value: 42}
		env.store["b"] = &Integer{Value: 73}

		a, ok := env.Get("a")
		if !ok {
			t.Error("expected 'a' to be in the environment")
		}
		aa := a.(*Integer)
		if aa.Value != 42 {
			t.Errorf("expected 'a' to be equal to %d, got %d", 42, aa.Value)
		}

		b, ok := env.Get("b")
		if !ok {
			t.Error("expected 'b' to be in the environment")
		}
		bb := b.(*Integer)
		if bb.Value != 73 {
			t.Errorf("expected 'b' to be equal to %d, got %d", 73, bb.Value)
		}

		c, ok := env.Get("c")
		if ok {
			t.Error("expected 'c' NOT to be in the environment")
		}
		if c != nil {
			t.Errorf("expected 'c' to be nil, got %v", c)
		}
	})

	t.Run("with one outer env", func(t *testing.T) {
		outer := NewEnvironment()
		outer.store["a"] = &Integer{Value: 1}
		env := NewEnclosedEnvironment(outer)
		env.store["b"] = &Integer{Value: 2}

		a, ok := env.Get("a")
		if !ok {
			t.Error("expected 'a' to be in the environment")
		}
		aa := a.(*Integer)
		if aa.Value != 1 {
			t.Errorf("expected 'a' to be equal to %d, got %d", 1, aa.Value)
		}

		b, ok := env.Get("b")
		if !ok {
			t.Error("expected 'b' to be in the environment")
		}
		bb := b.(*Integer)
		if bb.Value != 2 {
			t.Errorf("expected 'b' to be equal to %d, got %d", 2, bb.Value)
		}

		outerB, ok := outer.Get("b")
		if ok {
			t.Error("expected 'b' NOT to be in the outer environment")
		}
		if outerB != nil {
			t.Errorf("expected 'b' to be nil, got %v", outerB)
		}
	})
}

func TestEnvironment_Add(t *testing.T) {
	t.Run("simple add", func(t *testing.T) {
		env := NewEnvironment()

		if _, ok := env.store["a"]; ok {
			t.Error("expected 'a' NOT to be in the environment")
		}

		ok := env.Add("a", &Integer{Value: 1})
		if !ok {
			t.Error("expected env.Add to return true")
		}

		a, ok := env.store["a"]
		if !ok {
			t.Error("expected 'a' to be in the environment")
		}
		aa := a.(*Integer)
		if aa.Value != 1 {
			t.Errorf("expected 'a' to be equal to %d, got %d", 1, aa.Value)
		}
	})

	t.Run("add already existing variable", func(t *testing.T) {
		env := NewEnvironment()
		env.store["a"] = &Integer{Value: 1}

		ok := env.Add("a", &Integer{Value: 2})
		if ok {
			t.Error("expected env.Add to return false")
		}

		a, ok := env.store["a"]
		if !ok {
			t.Error("expected 'a' to be in the environment")
		}
		aa := a.(*Integer)
		if aa.Value != 1 {
			t.Errorf("expected 'a' to be equal to %d, got %d", 1, aa.Value)
		}
	})

	t.Run("add already existing variable in outer", func(t *testing.T) {
		outer := NewEnvironment()
		outer.store["a"] = &Integer{Value: 1}
		env := NewEnclosedEnvironment(outer)

		ok := env.Add("a", &Integer{Value: 2})
		if !ok {
			t.Error("expected env.Add to return true")
		}

		a, ok := env.store["a"]
		if !ok {
			t.Error("expected 'a' to be in the environment")
		}
		aa := a.(*Integer)
		if aa.Value != 2 {
			t.Errorf("expected 'a' to be equal to %d, got %d", 2, aa.Value)
		}
	})
}

func TestEnvironment_Set(t *testing.T) {
	t.Run("simple set", func(t *testing.T) {
		env := NewEnvironment()
		env.store["a"] = &Integer{Value: 1}

		ok := env.Set("a", &Integer{Value: 2})
		if !ok {
			t.Error("expected env.Set to return true")
		}

		a, ok := env.store["a"]
		if !ok {
			t.Error("expected 'a' to be in the environment")
		}
		aa := a.(*Integer)
		if aa.Value != 2 {
			t.Errorf("expected 'a' to be equal to %d, got %d", 2, aa.Value)
		}
	})

	t.Run("set not existing variable", func(t *testing.T) {
		env := NewEnvironment()

		ok := env.Set("a", &Integer{Value: 1})
		if ok {
			t.Error("expected env.Set to return false")
		}

		_, ok = env.store["a"]
		if ok {
			t.Error("expected 'a' NOT to be in the environment")
		}
	})

	t.Run("set variable in outer", func(t *testing.T) {
		outer := NewEnvironment()
		outer.store["a"] = &Integer{Value: 1}
		env := NewEnclosedEnvironment(outer)

		ok := env.Set("a", &Integer{Value: 2})
		if !ok {
			t.Error("expected env.Add to return true")
		}

		a, ok := env.store["a"]
		if ok {
			t.Error("expected 'a' NOT to be in the environment")
		}

		a, ok = outer.store["a"]
		if !ok {
			t.Error("expected 'a' to be in the outer environment")
		}
		aa := a.(*Integer)
		if aa.Value != 2 {
			t.Errorf("expected 'a' to be equal to %d, got %d", 2, aa.Value)
		}
	})
}
