import {
    loadFixture,
} from "@nomicfoundation/hardhat-toolbox/network-helpers";
import { expect } from "chai";
import hre from "hardhat";

describe("ValidatorPool", function () {
    // We define a fixture to reuse the same setup in every test.
    // We use loadFixture to run this setup once, snapshot that state,
    // and reset Hardhat Network to that snapshot in every test.
    async function deployValidatorPool() {
        // Contracts are deployed using the first signer/account by default
        const [owner, candidate] = await hre.ethers.getSigners();

        const ValidatorPool = await hre.ethers.getContractFactory("ValidatorPool");
        const pool = await ValidatorPool.deploy(owner.address, 1);

        return { pool, owner, candidate };
    }

    describe("Deployment", function () {
        it("Should set the correct owner", async function () {
            const { pool, owner } = await loadFixture(deployValidatorPool);
            expect(await pool.owner()).to.equal(owner);
        });
        it("Should set the correct number of validator slots", async function () {
            const { pool } = await loadFixture(deployValidatorPool);
            expect(await pool.slots()).to.equal(1);
        });
    });

    describe("submitApplication", function () {
        /*
            * Application is added as pending if all conditions are met
            * Application is automatically rejected if candidate already submitted another application
            * Application is automatically rejected if candidate is already a validator
            * Application is automatically rejected if the validator pool max cap is reached
        */
        it("Should revert if candidate is not msg.sender", async function () {
            const { pool, owner, candidate } = await loadFixture(deployValidatorPool);
            // submit application for owner account, from candidate account
            await expect(pool.connect(candidate).submitApplication(owner.address, owner.address)).to.be.revertedWithCustomError(pool, "InvalidOperator")
        });
        it("Should revert if candidate has an existing application", async function () {
            const { pool, candidate } = await loadFixture(deployValidatorPool);
            // submit first application
            pool.connect(candidate).submitApplication(candidate.address, candidate.address)
            // submit second application
            await expect(pool.connect(candidate).submitApplication(candidate.address, candidate.address)).to.be.revertedWithCustomError(pool, "ApplicationExists")
        });
        it("Should revert if pool has no empty slots", async function () {
            const { pool, owner, candidate } = await loadFixture(deployValidatorPool);
            // max validators is set to 1 for testing purposes
            // submit application from candidate account
            pool.connect(candidate).submitApplication(candidate.address, candidate.address)
            // approve application from owner account
            pool.connect(owner).approveApplication(candidate.address)
            // pool is now full. submit application from owner account
            await expect(pool.connect(owner).submitApplication(owner.address, owner.address)).to.be.revertedWithCustomError(pool, "NoValidatorSlots")
        });
        it("Should emit an ApplicationSubmitted event", async function () {
            const { pool, candidate } = await loadFixture(deployValidatorPool);
            // submit a valid application from the candidate account
            await expect(pool.connect(candidate).submitApplication(candidate.address, candidate.address)).to.emit(pool, "ApplicationSubmitted")
        });
    });

    describe("approveApplication", function () {
        /*
            * Application is marked as approved
            * Candidate is promoted to a validator
            * The new validator participates in consensus
            * All validators have an equal validation power
        */
        it("Should revert if not owner", async function () {
            const { pool, candidate } = await loadFixture(deployValidatorPool);
            // try to approve an application from the candidate account
            await expect(pool.connect(candidate).approveApplication(candidate.address)).to.be.revertedWithCustomError(pool, "OwnableUnauthorizedAccount");
        });
        it("Should revert if candidate does not have a pending application", async function () {
            const { pool, owner, candidate } = await loadFixture(deployValidatorPool);
            // try to approve an application from the candidate account
            await expect(pool.connect(owner).approveApplication(candidate.address)).to.be.revertedWithCustomError(pool, "InvalidState");
        });
        it("Should revert if pool has no empty slots", async function () {
            const { pool, owner, candidate } = await loadFixture(deployValidatorPool);
            // max validators (slots) is set to 1 for testing purposes
            // submit application from candidate account
            await pool.connect(candidate).submitApplication(candidate.address, candidate.address);
            // submit application from owner account
            await pool.connect(owner).submitApplication(owner.address, owner.address);
            // approve candidates application from owner account
            await pool.connect(owner).approveApplication(candidate.address)
            // pool is now full, owner still has a pending application
            // try to approve owners application
            await expect(pool.connect(owner).approveApplication(owner.address)).to.be.revertedWithCustomError(pool, "NoValidatorSlots")
        });
        it("Should decrement slots by 1", async function () {
            const { pool, owner, candidate } = await loadFixture(deployValidatorPool);
            // submit application from candidate account
            await pool.connect(candidate).submitApplication(candidate.address, candidate.address);
            // approve candidates application from owner account
            await pool.connect(owner).approveApplication(candidate.address)
            expect(await pool.connect(owner).slots()).to.equal(0);
        });
        it("Should emit an ApplicationApproved event", async function () {
            const { pool, owner, candidate } = await loadFixture(deployValidatorPool);
            // submit a application from candidate account
            await pool.connect(candidate).submitApplication(candidate.address, candidate.address);
            // submit a valid approval for candidate application
            await expect(pool.connect(owner).approveApplication(candidate.address)).to.emit(pool, "ApplicationApproved").and.emit(pool, "ValidatorJoined");
        });
    });

    describe("kick", function () {
        /*
            * Kicked is validator removed from the pool
            * Kicked validator does no longer participate in consensus
            * All remaining validators have an equal validation power
        */
        it("Should revert if not owner", async function () {
            const { pool, candidate } = await loadFixture(deployValidatorPool);
            await expect(pool.connect(candidate).kick(candidate.address)).to.be.revertedWithCustomError(pool, "OwnableUnauthorizedAccount");
        });
        it("Should revert if operator is not an approved validator", async function () {
            const { pool, owner, candidate } = await loadFixture(deployValidatorPool);
            await expect(pool.connect(owner).kick(candidate.address)).to.be.revertedWithCustomError(pool, "InvalidState");
        });
        it("Should increment slots by 1", async function () {
            const { pool, owner, candidate } = await loadFixture(deployValidatorPool);
            // submit application from candidate account
            await pool.connect(candidate).submitApplication(candidate.address, candidate.address);
            // approve candidates application from owner account
            await pool.connect(owner).approveApplication(candidate.address)
            // kick candidate
            await pool.connect(owner).kick(candidate.address)
            expect(await pool.connect(owner).slots()).to.equal(1);
        });
        it("Should emit a ValidatorKicked event", async function () {
            const { pool, owner, candidate } = await loadFixture(deployValidatorPool);
            // submit application from candidate account
            await pool.connect(candidate).submitApplication(candidate.address, candidate.address);
            // approve candidates application from owner account
            await pool.connect(owner).approveApplication(candidate.address)
            // kick candidate from validators
            await expect(pool.connect(owner).kick(candidate.address)).to.emit(pool, "ValidatorKicked")
        });
    });

    describe("leave", function () {
        /*
            * Validator is removed from the pool
            * Validator does no longer participate in consensus
            * All remaining validators have an equal validation power
        */
        it("Should revert if operator is not an approved validator", async function () {
            const { pool, candidate } = await loadFixture(deployValidatorPool);
            await expect(pool.connect(candidate).leave()).to.be.revertedWithCustomError(pool, "InvalidState");
        });
        it("Should increment slots by 1", async function () {
            const { pool, owner, candidate } = await loadFixture(deployValidatorPool);
            // submit application from candidate account
            await pool.connect(candidate).submitApplication(candidate.address, candidate.address);
            // approve candidates application from owner account
            await pool.connect(owner).approveApplication(candidate.address)
            // call leave from the candidate account
            await pool.connect(candidate).leave()
            expect(await pool.connect(owner).slots()).to.equal(1);
        });
        it("Should emit a ValidatorLeft event", async function () {
            const { pool, owner, candidate } = await loadFixture(deployValidatorPool);
            // submit application from candidate account
            await pool.connect(candidate).submitApplication(candidate.address, candidate.address);
            // approve candidates application from owner account
            await pool.connect(owner).approveApplication(candidate.address)
            await expect(pool.connect(candidate).leave()).to.emit(pool, "ValidatorLeft")
        });
    });
});
