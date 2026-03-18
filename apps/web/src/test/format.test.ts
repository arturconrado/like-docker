import { commandLabel } from '../lib/format'

describe('format helpers', () => {
  it('concatena comando e argumentos', () => {
    expect(commandLabel('ls', ['-la'])).toBe('ls -la')
  })

  it('tolera args nulo sem lançar erro', () => {
    expect(commandLabel('/bin/sh', null)).toBe('/bin/sh')
  })

  it('tolera args ausente sem lançar erro', () => {
    expect(commandLabel('pwd')).toBe('pwd')
  })

  it('tolera args inválido sem quebrar renderização', () => {
    expect(commandLabel('ls', { invalid: true } as unknown as string[])).toBe('ls')
  })
})
