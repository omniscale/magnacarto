MSS = $(wildcard *.mss)

EXPECTED_XML = $(patsubst %.mss,%.expected.xml,$(MSS))
EXPECTED_MAP = $(patsubst %.mss,%.expected.map,$(MSS))

# RESULTS = $(patsubst %.mss,%.result.xml,$(MSS))
COMPARE_XML = $(patsubst %.mss,%.cmp.xml,$(MSS))
COMPARE_MAP = $(patsubst %.mss,%.cmp.map,$(MSS))

MARGNACARTO_BIN = ../../magnacarto

all: $(COMPARE_MAP) $(COMPARE_XML)

.PHONY: all refresh-expected

$(MARGNACARTO_BIN):
	(cd ../../ && go build ./cmd/magnacarto)

%.result.xml: %.mss $(MARGNACARTO_BIN)
	../../magnacarto -mss $< > $@

%.result.map: %.mss $(MARGNACARTO_BIN)
	../../magnacarto -mss $< -builder mapserver -ms-no-map-block > $@


%.cmp.xml: %.expected.xml %.result.xml
	diff -wu $^

%.cmp.map: %.expected.map %.result.map
	diff -wu $^

refresh-expected: $(EXPECTED_XML) $(EXPECTED_MAP)

%.expected.xml: %.mss $(MARGNACARTO_BIN)
	../../magnacarto -mss $< > $@

%.expected.map: %.mss $(MARGNACARTO_BIN)
	../../magnacarto -mss $< -builder mapserver -ms-no-map-block > $@

# clean:
