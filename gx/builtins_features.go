package gx

var NewProperties = &Builtin{
	Name: "newProps",
	NativeCalls: []*NativeCall{
		&NativeCall{
			Signature: &Signature{
				Accepts: []Value{},
				Returns: &Raw{},
			},
			Call: func(args ...Value) (Value, error) {
				newProps := make(map[string]string)
				return &Raw{newProps}, nil
			},
		},
	},
}

var DeleteProperty = &Builtin{
	Name: "delete",
	NativeCalls: []*NativeCall{
		&NativeCall{
			Signature: &Signature{
				Accepts: []Value{
					&Raw{},
					&Str{},
				},
				Returns: &Str{},
			},
			Call: func(args ...Value) (Value, error) {
				props := args[0].(*Raw).NativeValue.(map[string]string)
				key := args[1].(*Str).NativeValue
				// TODO: Should we throw an error if the key doesn't exist?
				oldValue := &Str{props[key]}
				delete(props, key)
				return oldValue, nil
			},
		},
	},
}
