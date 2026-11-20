Feature: Withdrawing cash
  Como cliente autenticado
  Quiero retirar efectivo indicando un monto
  Para disponer de mis ahorros y ver mi nuevo saldo

  Rule: Customers cannot withdraw more than their balance

    Scenario: Successful withdrawal within balance
      Given Alice has 234.56 in their account
      When Alice tries to withdraw 200.00
      Then the withdrawal is successful
      And the balance is 34.56

    Scenario: Declined withdrawal in excess of balance
      Given Hamza has 198.76 in their account
      When Hamza tries to withdraw 200.00
      Then the withdrawal is declined
      And the balance is 198.76

    Scenario: Invalid withdrawal amount
      Given Pedro has 300.00 in their account
      When Pedro tries to withdraw -20.00
      Then the system shows a validation error
      And the balance is 300.00
