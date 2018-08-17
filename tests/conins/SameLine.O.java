/** Test for same-line edits. */
package com.wesalvaro.test;

public class SameLine {
    public static final int A = 1;
    public static final int B = 2;
<<<<<<< LOCAL
    public static final int C = 3;
    public static final int E = 4;
=======
    public static final int D = 3;
>>>>>>> OTHER
    public ErrorMsg( ) {
        ema = new String[10];
        ema[A] = ReBu.grbGet("A.text");
        ema[B] = ReBu.grbGet("B.text");
        ema[C] = ReBu.grbGet("text_fileOne.text");
        ema[D] = ReBu.grbGet("plain_fileFoo.text");
    }
}

