# US-01b: Abono en efectivo
# As a:    cliente autenticado
# I want:  abonar (depositar) efectivo en mi cuenta
# so that: incrementar mi saldo disponible
#
# Acceptance criteria:
#   Dado que estoy autenticado y mi cuenta está activa, cuando ingreso un monto
#   válido mayor a 0, el sistema suma el monto al saldo y muestra el nuevo saldo.
#   Si el monto es cero o negativo, el sistema muestra un error y no cambia el saldo.

Feature: Depositing cash

  Rule: Deposits must be greater than zero

    Scenario: Successful cash deposit
      Given Alice has 100.00 in their account
      When Alice deposits 50.50
      Then the deposit is successful
      And the balance is 150.50

    Scenario: Rejected deposit with zero amount
      Given Hamza has 80.00 in their account
      When Hamza deposits 0.00
      Then the system shows a validation error
      And the balance is 80.00

    Scenario: Rejected deposit with negative amount
      Given Pedro has 80.00 in their account
      When Pedro deposits -10.00
      Then the system shows a validation error
      And the balance is 80.00
