foo = 12;
bar = foo * foo;
if (bar == 144)
    foo = 10;
if (bar <= foo*foo)
    return 0;
else
    if (foo == 10)
        return bar - foo;