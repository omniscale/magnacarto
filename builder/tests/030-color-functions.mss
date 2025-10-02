#l {
    line-width: 1;
}

@base: #a61;

#l [func="saturate"] { line-color: saturate(@base, 20); }
#l [func="desaturate"] { line-color: desaturate(@base, 20); }
#l [func="fadein"] { line-color: fadein(@base, 20); }
#l [func="fadeout"] { line-color: fadeout(@base, 20); }
#l [func="spin"] { line-color: spin(@base, 20); }
#l [func="lighten"] { line-color: lighten(@base, 20); }
#l [func="darken"] { line-color: darken(@base, 20); }


@rgba: rgba(100, 20, 180, 0.5);

#l [func="saturate_rgba"] { line-color: saturate(@rgba, 20); }
#l [func="desaturate_rgba"] { line-color: desaturate(@rgba, 20); }
#l [func="fadein_rgba"] { line-color: fadein(@rgba, 20); }
#l [func="fadeout_rgba"] { line-color: fadeout(@rgba, 20); }
#l [func="spin_rgba"] { line-color: spin(@rgba, 20); }
#l [func="lighten_rgba"] { line-color: lighten(@rgba, 20); }
#l [func="darken_rgba"] { line-color: darken(@rgba, 20); }


#l [func="null_grey"] { line-color: grey; }
#l [func="saturate_grey"] { line-color: saturate(grey, 20); }
#l [func="desaturate_grey"] { line-color: desaturate(grey, 20); }
#l [func="fadein_grey"] { line-color: fadein(grey, 20); }
#l [func="fadeout_grey"] { line-color: fadeout(grey, 20); }
#l [func="spin_grey"] { line-color: spin(grey, 20); }
#l [func="lighten_grey"] { line-color: lighten(grey, 20); }
#l [func="darken_grey"] { line-color: darken(grey, 20); }
