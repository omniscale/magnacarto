#test {
  [id=1] {
    [zoom >= 13] {
      line-width: 1;
      line-color: black;
      [name != 'foo'] {
        [id=1] { line-width: 4; }
      }
    }
    [zoom >= 14] {
      line-width: 99;
    }
    [zoom >= 15] {
      line-color: white;
      line-width: 20;
    }
    [zoom >= 17] {
      line-width: 99;
    }
  }
 }




// .roads-fill[zoom >= 10] {
//   ::fill {

//     [feature = 'highway_footway'],
//     [feature = 'highway_path'][foot = 'designated'] {
//       [zoom >= 13][access != 'no'],
//       [zoom >= 15] {
//         line/line-color: @footway-fill;
//         line/line-width: @footway-width-z13;
//         [zoom >= 15] { line/line-width:  @footway-width-z15; }
//       }
//     }

//     [feature = 'highway_cycleway'],
//     [feature = 'highway_path'][bicycle = 'designated'] {
//       [zoom >= 13][access != 'no'],
//       [zoom >= 15] {
//         line/line-color: @cycleway-fill;
//         line/line-width: @cycleway-width-z13;
//         [zoom >= 15] { line/line-width: 3.3; } // foot=designated, bicycle=designated should use this
//       }
//     }
// }
// }