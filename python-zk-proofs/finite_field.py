from finitefield.modp import IntegersModP
if __name__ == "__main__":
    mod7 = IntegersModP(7)
    ret = mod7(3) + mod7(6)
    # 9 mod 7 = 2
    print(ret)

    mod23 = IntegersModP(23)
    mod23_7_inverse = mod23(7).inverse()
    ret = mod23(7) * mod23_7_inverse
    print(mod23_7_inverse, ret)
