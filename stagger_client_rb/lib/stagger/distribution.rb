class Distribution
  attr_reader :weight, :sum_x, :sum_x2, :min, :max

  def initialize
    @weight, @sum_x, @sum_x2, @min, @max = 0, 0, 0, nil, nil
  end

  def add(x, weight = 1)
    @weight += weight
    @sum_x += x * weight
    @sum_x2 += x**2 * weight
    @min = [@min, x].compact.min
    @max = [@max, x].compact.max
    self
  end

  def mean
    @weight > 0 ? @sum_x.to_f / @weight : nil
  end

  def to_a
    [@weight, @min, @max, @sum_x, @sum_x2]
  end
end
