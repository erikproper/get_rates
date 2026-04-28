/*
 *
 * Module:    GetRates
 * Package:   Config
 * Component: Funds
 *
 * Defines the complete list of funds to be scraped, grouped by data source.
 *
 * Creator: Henderik A. Proper (e.proper@acm.org), Luxembourg, in collaboration with Claude.ai
 *
 * Version of: 28.04.2026
 *
 */

package config

import "get_rates/internal/model"

// Funds is the ordered list of all funds to scrape and push to the sheet.
var Funds = []model.TFund{

	// --- Tradegate ---

	{ISIN: "IE00B4ND3602", Name: "iShares Physical Gold ETC", Source: model.SourceTradegate},
	{ISIN: "IE00B4NCWG09", Name: "iShares Physical Silver ETC", Source: model.SourceTradegate},
	{ISIN: "NL0000337319", Name: "Koninklijke BAM Groep NV", Source: model.SourceTradegate},
	{ISIN: "IE000U9ODG19", Name: "iShares Global Aerospace & Defence ETF", Source: model.SourceTradegate},
	{ISIN: "IE00B4L5Y983", Name: "iShares Core MSCI World UCITS ETF", Source: model.SourceTradegate},
	{ISIN: "DE0007030009", Name: "Rheinmetall AG", Source: model.SourceTradegate},
	{ISIN: "BMG3602E1084", Name: "Flow Traders Ltd", Source: model.SourceTradegate},
	{ISIN: "IE000YYE6WK5", Name: "VanEck Defense UCITS ETF A", Source: model.SourceTradegate},
	{ISIN: "IE00B5L8K969", Name: "iShares MSCI EM Asia UCITS ETF", Source: model.SourceTradegate},
	{ISIN: "IE00BK5BC891", Name: "L&G Clean Water UCITS ETF", Source: model.SourceTradegate},
	{ISIN: "JE00BQRFDY49", Name: "WisdomTree Core Physical Silver ETC", Source: model.SourceTradegate},
	{ISIN: "JE00BN2CJ301", Name: "WisdomTree Core Physical Gold ETC", Source: model.SourceTradegate},
	{ISIN: "DE000A0H08F7", Name: "iShares STOXX Europe 600 Construction & Materials ETF", Source: model.SourceTradegate},
	{ISIN: "IE0006WW1TQ4", Name: "Xtrackers MSCI World ex USA UCITS ETF 1C", Source: model.SourceTradegate},
	{ISIN: "IE00BD1F4M44", Name: "iShares Edge MSCI USA Value Factor UCITS ETF", Source: model.SourceTradegate},
	{ISIN: "IE00B5BMR087", Name: "iShares Core S&P 500 UCITS ETF", Source: model.SourceTradegate},
	{ISIN: "IE000CFH1JX2", Name: "iShares Global Water UCITS ETF", Source: model.SourceTradegate},
	{ISIN: "IE00BYTRRB94", Name: "SPDR MSCI World Health Care UCITS ETF", Source: model.SourceTradegate},
	{ISIN: "IE00BM67HK77", Name: "Xtrackers MSCI World Health Care UCITS ETF 1C", Source: model.SourceTradegate},
	{ISIN: "IE00BMH5Y327", Name: "Global X Data Center REITS & Digital Infrastructure ETF", Source: model.SourceTradegate},
	{ISIN: "IE000CK5G8J7", Name: "iShares Global Infrastructure UCITS ETF", Source: model.SourceTradegate},
	{ISIN: "IE00BYZK4552", Name: "iShares Automation & Robotics UCITS ETF", Source: model.SourceTradegate},
	{ISIN: "IE00B53SZB19", Name: "iShares Nasdaq 100 UCITS ETF", Source: model.SourceTradegate},
	{ISIN: "IE000E7EI9P0", Name: "Amundi S&P Global Information Technology ESG UCITS ETF", Source: model.SourceTradegate},
	{ISIN: "IE00BF5DXP42", Name: "First Trust Index Innovative Transaction & Process UCITS ETF", Source: model.SourceTradegate},

	// --- FT Markets ---

	{ISIN: "NL0012706154", Name: "BND Wereld Indexfonds Hedged", Source: model.SourceFTMarkets},
	{ISIN: "NL0012706196", Name: "BND Wereld Obligatie Indexfonds", Source: model.SourceFTMarkets},
	{ISIN: "NL0012706238", Name: "BND Wereld Indexfonds Unhedged", Source: model.SourceFTMarkets},
	{ISIN: "NL0014040164", Name: "BND Emerging Markets Indexfonds", Source: model.SourceFTMarkets},
	{ISIN: "NL0013546955", Name: "BND Small Cap Wereld Indexfonds Carb Screened", Source: model.SourceFTMarkets},
	{ISIN: "NL0012706220", Name: "BND Euro Investment Grade Obligatie Indexfonds", Source: model.SourceFTMarkets},
	{ISIN: "NL0012706204", Name: "BND Euro Staatsobligatie Indexfonds", Source: model.SourceFTMarkets},
	{ISIN: "NL0012706279", Name: "BND Euro Staatsobligatie Indexfonds Inflatie", Source: model.SourceFTMarkets},
	{ISIN: "IE00BVVQ9K69", Name: "Selected Screened FTSE Developed World II (B) Common Contractual Fund - EUR Hedged Acc", Source: model.SourceFTMarkets},
	{ISIN: "IE000FZ5BIF1", Name: "Vanguard Global Government Bond Index EUR Hedged Acc", Source: model.SourceFTMarkets},
	{ISIN: "IE00BNDS0V25", Name: "Vanguard ESG Global Corporate Bond Index Fund EUR Hedged Acc", Source: model.SourceFTMarkets},
	{ISIN: "IE00BVVQBD33", Name: "Vanguard Selected Screened FTSE Developed II (B) World Common Contractual Fd Instl B EUR Acc", Source: model.SourceFTMarkets},
	{ISIN: "NL0013089147", Name: "Northern Trust UCITS FGR Fund - Emerging Markets Screened Equity Index FGR Fund E EUR", Source: model.SourceFTMarkets},
	{ISIN: "NL0013552094", Name: "NT World Small Cap Low Carbon Equity Index FGR Feeder Fund Class C EUR Units", Source: model.SourceFTMarkets},
	{ISIN: "IE00BYSX5D68", Name: "Vanguard Selected Screened Euro Investment Grade Bond Index Fund EUR Acc", Source: model.SourceFTMarkets},
	{ISIN: "IE0007472990", Name: "Vanguard Euro Government Bond Index Fund EUR Acc", Source: model.SourceFTMarkets},
}
