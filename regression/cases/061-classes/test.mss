.testlines {
    line-width: 5;
}

.testlines {
    [id=1] {
        line-width: 1;
    }
}

[id=2] {
    .testlines {
        line-width: 2;
    }
}

#test[id=3] {
    .testlines {
        line-color: red;
    }
    line-width: 3;
}
