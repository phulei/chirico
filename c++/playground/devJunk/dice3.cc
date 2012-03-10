#include <iostream>
using std::cout;
using std::endl;

#include <cstdlib>
using std::rand;
using std::srand;

#include <map>

typedef std::map < int, int >IntMap;


#include <ctime>
using std::time;


int rollDice();
class Stuff;

int rollDice()
{
  int die1 = 1 + rand() % 6;
  int die2 = 1 + rand() % 6;

  int sum = die1 + die2;

  return sum;
}

class Stuff {
public:
  Stuff():die1(0),die2(0){}


private:
  int die1;
  int die2;

};




int main(void)
{

  enum Status { CONTINUE, WON, LOST };
  IntMap m;
  Status gameStatus;
  int count = 0;
  srand(time(0));
  int sumOfDice = rollDice();

  for (int i = 2; i <= 12; ++i)
    m[i] = 0;


  for (int i = 0; i < 2000; ++i)
    {
      m[sumOfDice] += 1;
      sumOfDice = rollDice();
    }

  for (int i = 2; i <= 12; ++i)
    printf("m[%2d]=%8d\n", i, m[i]);



  return 0;

}


