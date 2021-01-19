a = 3;
prime = a;
while (a < 100) {
  flag = 1;
  i = 2;
  while (i*i <= a) {
    if (a%i == 0) {
      flag = 0;
    }
    i = i+1;
  }
  if (flag) {
    prime = a;
  }
  a = a+1;
}
return prime;