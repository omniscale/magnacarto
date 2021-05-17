#class[zoom=2] {
    line-width: 1;
}

#class_1_2[zoom=3][foo != 42] {
    line-width: 1;
}


#class_1_2 [zoom=3]  ["foo:bar" != 42] {
    line-width: 1;
}

#class['quoted'="bar"],
#class['quoted:quoted'="bar"],
#class["quoted2:quoted"='bar'] {
    line-width: 2;
}

#class_3 [zoom=3] {
    [type='foo'], [type='bar'], [type='baz'] {
        line-width: 1;
    }
    [class="foo"], [class="bar"], [class="baz"] {
        line-color: #f00;
    }
}

#class_4[id % 10 = 0] {
    line-width: 2;
}
