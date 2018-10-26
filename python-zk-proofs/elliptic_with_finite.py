
from elliptic_basic import EllipticCurve, Point
from finitefield.finitefield import FiniteField

# as we have known, finite field using euclidean for finding inverse
if __name__ == "__main__":
    F5 = FiniteField(5, 1)
    C = EllipticCurve(a=F5(1), b=F5(1))
    P = Point(C, F5(2), F5(1))

    print(P, 2*P, 3*P)

    #  fieldsize 25
    F25 = FiniteField(5, 2)
    print(F25.idealGenerator)

    curve = EllipticCurve(a=F25([1]), b=F25([1]))

    x = F25([2, 1])
    y = F25([0, 2])
    print(y*y - x*x*x - x - 1)
    print(curve.testPoint(x, y))

    P = Point(curve, F25([2]), F25([1]))
    print(-P, P+P, 4*P, 9*P)
