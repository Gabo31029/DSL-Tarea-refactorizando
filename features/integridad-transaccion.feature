# US-04: Integridad de la transacción
# As a:    banco (servicio ATM)
# I want:  que cada retiro/abono sea atómico y quede registrado
# so that: garantizar consistencia contable y auditoría
#
# Acceptance criteria:
#   Una transacción se completa por entero o no se hace (RN-03).
#   Todo intento deja rastro en el log JSON con estado EXITOSA, RECHAZADA o ERROR.

Feature: Transaction integrity

  Rule: Every transaction is atomic and logged

    Background:
      Given the accounts are loaded from "data/cuentas.json"

    Scenario: Successful transaction is logged
      Given account C001 has 234.56
      When a RETIRO of 200.00 is processed on C001
      Then the balance of C001 is 34.56
      And the log has a record with estado "EXITOSA"

    Scenario: Declined transaction is logged and balance unchanged
      Given account C002 has 198.76
      When a RETIRO of 200.00 is processed on C002
      Then the balance of C002 is 198.76
      And the log has a record with estado "RECHAZADA"

    Scenario: Error transaction is logged and balance unchanged
      Given account C003 has 300.00
      When an ABONO of -10.00 is processed on C003
      Then the balance of C003 is 300.00
      And the log has a record with estado "ERROR"

    Scenario: Inactive or unknown account
      When a RETIRO of 10.00 is processed on C004
      Then the log has a record with estado "ERROR"
