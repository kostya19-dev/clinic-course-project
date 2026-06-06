LATEX ?= xelatex
GS ?= gs

MAIN := main
PDF := $(MAIN).pdf
COMPRESSED_PDF := $(MAIN)_compressed.pdf

.PHONY: all report compress clean-report

all: report

report:
	$(LATEX) -interaction=batchmode $(MAIN).tex
	$(LATEX) -interaction=batchmode $(MAIN).tex
	$(MAKE) compress

compress: $(PDF)
	$(GS) -sDEVICE=pdfwrite \
		-dCompatibilityLevel=1.4 \
		-dNOPAUSE \
		-dOptimize=true \
		-dQUIET \
		-dBATCH \
		-dRemoveUnusedFonts=true \
		-dRemoveUnusedImages=true \
		-dOptimizeResources=true \
		-dDetectDuplicateImages \
		-dCompressFonts=true \
		-dEmbedAllFonts=true \
		-dSubsetFonts=true \
		-dPreserveAnnots=true \
		-dPreserveMarkedContent=true \
		-dPreserveOverprintSettings=true \
		-dPreserveHalftoneInfo=true \
		-dPreserveOPIComments=true \
		-dPreserveDeviceN=true \
		-dMaxInlineImageSize=0 \
		-sOutputFile=$(COMPRESSED_PDF) \
		$(PDF)
	mv $(COMPRESSED_PDF) $(PDF)

clean-report:
	rm -f *.aux *.log *.toc *.lof *.lot *.out *.synctex.gz *.bbl *.blg *.fls *.fdb_latexmk $(PDF) $(COMPRESSED_PDF)
