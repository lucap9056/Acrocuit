declare const brandSymbol: unique symbol;
type Brand<T, B> = T & { readonly [brandSymbol]: B };
